package handler

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"

	papi "github.com/redhat-cop/podpreset-webhook/pkg/apis/redhatcop/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
	at "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

const (
	podpresetName = "podpresets.admission.redhatcop.redhat.io"
)

var log = logf.Log.WithName("podpreset_webhook_handler")

type PodPresetMutator struct {
	client  client.Client
	decoder at.Decoder
}

func (p *PodPresetMutator) Handle(ctx context.Context, req at.Request) at.Response {

	log.V(4).Info("Webhook Invoked", "Request", req.AdmissionRequest)
	pod := &corev1.Pod{}

	err := p.decoder.Decode(req, pod)
	if err != nil {
		log.Error(err, "Error occurred Decoding Pod")
		return admission.ErrorResponse(http.StatusBadRequest, err)
	}
	copy := pod.DeepCopy()

	err = p.mutatePodsFn(ctx, copy)

	if err != nil {
		log.Error(err, "Error occurred mutating Pod")
		return admission.ErrorResponse(http.StatusInternalServerError, err)
	}

	// admission.PatchResponse generates a Response containing patches.
	return admission.PatchResponse(pod, copy)

}

// Mutates function values
func (p *PodPresetMutator) mutatePodsFn(ctx context.Context, pod *corev1.Pod) error {

	if _, isMirrorPod := pod.Annotations[corev1.MirrorPodAnnotationKey]; isMirrorPod {
		return nil
	}

	// Ignore if exclusion annotation is present
	if podAnnotations := pod.GetAnnotations(); podAnnotations != nil {
		log.Info("Looking at pod annotations", "found", podAnnotations)
		if podAnnotations[corev1.PodPresetOptOutAnnotationKey] == "true" {
			return nil
		}
	}

	podPresetList := &papi.PodPresetList{}

	err := p.client.List(context.TODO(), &client.ListOptions{}, podPresetList)

	if err != nil {
		return fmt.Errorf("listing pod presets failed: %v", err)
	}

	matchingPPs, err := filterPodPresets(podPresetList, pod)
	if err != nil {
		return fmt.Errorf("filtering pod presets failed: %v", err)
	}

	if len(matchingPPs) == 0 {
		return nil
	}

	presetNames := make([]string, len(matchingPPs))
	for i, pp := range matchingPPs {
		presetNames[i] = pp.GetName()
	}

	// detect merge conflict
	err = safeToApplyPodPresetsOnPod(pod, matchingPPs)
	if err != nil {
		// conflict, ignore the error, but raise an event
		log.Info("conflict occurred while applying podpresets: %s on pod", "Names", strings.Join(presetNames, ","), "Pod Name", pod.GetGenerateName())
		return nil
	}

	applyPodPresetsOnPod(pod, matchingPPs)

	log.Info("applied podpresets", "Names", strings.Join(presetNames, ","), "Pod Name", pod.GetGenerateName())

	return nil
}

// applyPodPresetsOnPod updates the PodSpec with merged information from all the
// applicable PodPresets. It ignores the errors of merge functions because merge
// errors have already been checked in safeToApplyPodPresetsOnPod function.
func applyPodPresetsOnPod(pod *corev1.Pod, podPresets []*papi.PodPreset) {
	if len(podPresets) == 0 {
		return
	}

	volumes, _ := mergeVolumes(pod.Spec.Volumes, podPresets)
	pod.Spec.Volumes = volumes

	for i, ctr := range pod.Spec.Containers {
		applyPodPresetsOnContainer(&ctr, podPresets)
		pod.Spec.Containers[i] = ctr
	}

	// add annotation
	if pod.ObjectMeta.Annotations == nil {
		pod.ObjectMeta.Annotations = map[string]string{}
	}

	for _, pp := range podPresets {
		pod.ObjectMeta.Annotations[fmt.Sprintf("%s/podpreset-%s", podpresetName, pp.GetName())] = pp.GetResourceVersion()
	}
}

// applyPodPresetsOnContainer injects envVars, VolumeMounts and envFrom from
// given podPresets in to the given container. It ignores conflict errors
// because it assumes those have been checked already by the caller.
func applyPodPresetsOnContainer(ctr *corev1.Container, podPresets []*papi.PodPreset) {
	envVars, _ := mergeEnv(ctr.Env, podPresets)
	ctr.Env = envVars

	volumeMounts, _ := mergeVolumeMounts(ctr.VolumeMounts, podPresets)
	ctr.VolumeMounts = volumeMounts

	envFrom, _ := mergeEnvFrom(ctr.EnvFrom, podPresets)
	ctr.EnvFrom = envFrom
}

// filterPodPresets returns list of PodPresets which match given Pod.
func filterPodPresets(list *papi.PodPresetList, pod *corev1.Pod) ([]*papi.PodPreset, error) {
	var matchingPPs []*papi.PodPreset

	for i, pp := range list.Items {
		selector, err := metav1.LabelSelectorAsSelector(&pp.Spec.Selector)
		if err != nil {
			return nil, fmt.Errorf("label selector conversion failed: %v for selector: %v", pp.Spec.Selector, err)
		}

		// check if the pod labels match the selector
		if !selector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		log.Info("PodPreset matches pod labels", "PodPreset", pp.GetName(), "Pod", pod.GetName())
		matchingPPs = append(matchingPPs, &list.Items[i])
	}
	return matchingPPs, nil
}

// safeToApplyPodPresetsOnPod determines if there is any conflict in information
// injected by given PodPresets in the Pod.
func safeToApplyPodPresetsOnPod(pod *corev1.Pod, podPresets []*papi.PodPreset) error {
	var errs []error

	// volumes attribute is defined at the Pod level, so determine if volumes
	// injection is causing any conflict.
	if _, err := mergeVolumes(pod.Spec.Volumes, podPresets); err != nil {
		errs = append(errs, err)
	}
	for _, ctr := range pod.Spec.Containers {
		if err := safeToApplyPodPresetsOnContainer(&ctr, podPresets); err != nil {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}

// mergeVolumes merges given list of Volumes with the volumes injected by given
// podPresets. It returns an error if it detects any conflict during the merge.
func mergeVolumes(volumes []corev1.Volume, podPresets []*papi.PodPreset) ([]corev1.Volume, error) {
	origVolumes := map[string]corev1.Volume{}
	for _, v := range volumes {
		origVolumes[v.Name] = v
	}

	mergedVolumes := make([]corev1.Volume, len(volumes))
	copy(mergedVolumes, volumes)

	var errs []error

	for _, pp := range podPresets {
		for _, v := range pp.Spec.Volumes {
			found, ok := origVolumes[v.Name]
			if !ok {
				// if we don't already have it append it and continue
				origVolumes[v.Name] = v
				mergedVolumes = append(mergedVolumes, v)
				continue
			}

			// make sure they are identical or throw an error
			if !reflect.DeepEqual(found, v) {
				errs = append(errs, fmt.Errorf("merging volumes for %s has a conflict on %s: \n%#v\ndoes not match\n%#v\n in container", pp.GetName(), v.Name, v, found))
			}
		}
	}

	err := utilerrors.NewAggregate(errs)
	if err != nil {
		return nil, err
	}

	if len(mergedVolumes) == 0 {
		return nil, nil
	}

	return mergedVolumes, err
}

// safeToApplyPodPresetsOnContainer determines if there is any conflict in
// information injected by given PodPresets in the given container.
func safeToApplyPodPresetsOnContainer(ctr *corev1.Container, podPresets []*papi.PodPreset) error {
	var errs []error
	// check if it is safe to merge env vars and volume mounts from given podpresets and
	// container's existing env vars.
	if _, err := mergeEnv(ctr.Env, podPresets); err != nil {
		errs = append(errs, err)
	}
	if _, err := mergeVolumeMounts(ctr.VolumeMounts, podPresets); err != nil {
		errs = append(errs, err)
	}

	return utilerrors.NewAggregate(errs)
}

// mergeEnv merges a list of env vars with the env vars injected by given list podPresets.
// It returns an error if it detects any conflict during the merge.
func mergeEnv(envVars []corev1.EnvVar, podPresets []*papi.PodPreset) ([]corev1.EnvVar, error) {
	origEnv := map[string]corev1.EnvVar{}
	for _, v := range envVars {
		origEnv[v.Name] = v
	}

	mergedEnv := make([]corev1.EnvVar, len(envVars))
	copy(mergedEnv, envVars)

	var errs []error

	for _, pp := range podPresets {
		for _, v := range pp.Spec.Env {

			found, ok := origEnv[v.Name]
			if !ok {
				// if we don't already have it append it and continue
				origEnv[v.Name] = v
				mergedEnv = append(mergedEnv, v)
				continue
			}

			// make sure they are identical or throw an error
			if !reflect.DeepEqual(found, v) {
				errs = append(errs, fmt.Errorf("merging env for %s has a conflict on %s: \n%#v\ndoes not match\n%#v\n in container", pp.GetName(), v.Name, v, found))
			}
		}
	}

	err := utilerrors.NewAggregate(errs)
	if err != nil {
		return nil, err
	}

	return mergedEnv, err
}

func mergeEnvFrom(envSources []corev1.EnvFromSource, podPresets []*papi.PodPreset) ([]corev1.EnvFromSource, error) {
	var mergedEnvFrom []corev1.EnvFromSource

	mergedEnvFrom = append(mergedEnvFrom, envSources...)
	for _, pp := range podPresets {
		for _, envFromSource := range pp.Spec.EnvFrom {
			// internalEnvFrom := api.EnvFromSource{}
			// if err := apiscorev1.Convert_v1_EnvFromSource_To_core_EnvFromSource(&envFromSource, &internalEnvFrom, nil); err != nil {
			// 	return nil, err
			// }
			mergedEnvFrom = append(mergedEnvFrom, envFromSource)
		}

	}

	return mergedEnvFrom, nil
}

// mergeVolumeMounts merges given list of VolumeMounts with the volumeMounts
// injected by given podPresets. It returns an error if it detects any conflict during the merge.
func mergeVolumeMounts(volumeMounts []corev1.VolumeMount, podPresets []*papi.PodPreset) ([]corev1.VolumeMount, error) {

	origVolumeMounts := map[string]corev1.VolumeMount{}
	volumeMountsByPath := map[string]corev1.VolumeMount{}
	for _, v := range volumeMounts {
		origVolumeMounts[v.Name] = v
		volumeMountsByPath[v.MountPath] = v
	}

	mergedVolumeMounts := make([]corev1.VolumeMount, len(volumeMounts))
	copy(mergedVolumeMounts, volumeMounts)

	var errs []error

	for _, pp := range podPresets {
		for _, v := range pp.Spec.VolumeMounts {
			found, ok := origVolumeMounts[v.Name]
			if !ok {
				// if we don't already have it append it and continue
				origVolumeMounts[v.Name] = v
				mergedVolumeMounts = append(mergedVolumeMounts, v)
			} else {
				// make sure they are identical or throw an error
				// shall we throw an error for identical volumeMounts ?
				if !reflect.DeepEqual(found, v) {
					errs = append(errs, fmt.Errorf("merging volume mounts for %s has a conflict on %s: \n%#v\ndoes not match\n%#v\n in container", pp.GetName(), v.Name, v, found))
				}
			}

			found, ok = volumeMountsByPath[v.MountPath]
			if !ok {
				// if we don't already have it append it and continue
				volumeMountsByPath[v.MountPath] = v
			} else {
				// make sure they are identical or throw an error
				if !reflect.DeepEqual(found, v) {
					errs = append(errs, fmt.Errorf("merging volume mounts for %s has a conflict on mount path %s: \n%#v\ndoes not match\n%#v\n in container", pp.GetName(), v.MountPath, v, found))
				}
			}
		}
	}

	err := utilerrors.NewAggregate(errs)
	if err != nil {
		return nil, err
	}

	return mergedVolumeMounts, err
}

// PodPresetMutator implements inject.Client.
var _ inject.Client = &PodPresetMutator{}

// InjectClient injects the client into the PodAnnotator
func (a *PodPresetMutator) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// PodPresetMutator implements inject.Decoder.
var _ inject.Decoder = &PodPresetMutator{}

// InjectDecoder injects the decoder into the PodPresetMutator
func (a *PodPresetMutator) InjectDecoder(d at.Decoder) error {
	a.decoder = d
	return nil
}
