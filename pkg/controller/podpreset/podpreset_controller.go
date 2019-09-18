package podpreset

import (
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/redhat-cop/podpreset-webhook/pkg/controller/podpreset/handler"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission/builder"
	at "sigs.k8s.io/controller-runtime/pkg/webhook/admission/types"
)

const (
	webhookName             = "podpreset-webhook"
	webhookSecretName       = "podpreset-webhook-cert"
	serverPort        int32 = 8443
	certDir                 = "/tmp/cert"
	podpresetName           = "podpresets.admission.redhatcop.redhat.io"
)

var log = logf.Log.WithName("controller_podpreset")

type podPresetMutatingController struct {
	client  client.Client
	decoder at.Decoder
}

func Add(mgr manager.Manager) error {

	ns, err := k8sutil.GetWatchNamespace()

	if err != nil {
		return err
	}

	mutatingWebhook, err := builder.NewWebhookBuilder().
		Name(podpresetName).
		Mutating().
		Operations(admissionregistrationv1beta1.Create).
		WithManager(mgr).
		ForType(&corev1.Pod{}).
		FailurePolicy(admissionregistrationv1beta1.Ignore).
		Handlers(&handler.PodPresetMutator{}).
		Build()

	if err != nil {
		log.Error(err, "Error occurred building mutating webhook")
		return err
	}

	svr, err := webhook.NewServer(webhookName, mgr, webhook.ServerOptions{
		Port:    serverPort,
		CertDir: certDir,
		BootstrapOptions: &webhook.BootstrapOptions{
			Secret: &types.NamespacedName{
				Namespace: ns,
				Name:      webhookSecretName,
			},

			Service: &webhook.Service{
				Namespace: ns,
				Name:      webhookName,
				Selectors: map[string]string{
					"name": webhookName,
				},
			},
		},
	})

	if err != nil {
		log.Error(err, "Error occurred creating webhook server")
		return err
	}

	svr.Register(mutatingWebhook)

	return nil
}
