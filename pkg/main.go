package pkg

import (
	"github.com/pkg/errors"
	"github.com/plantoncloud/openfga-kubernetes-pulumi-module/pkg/outputs"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/openfgakubernetes"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/pulumikubernetesprovider"
	kubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *openfgakubernetes.OpenfgaKubernetesStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	//create kubernetes-provider from the credential in the stack-input
	kubernetesProvider, err := pulumikubernetesprovider.GetWithKubernetesClusterCredential(ctx,
		stackInput.KubernetesCluster, "kubernetes")
	if err != nil {
		return errors.Wrap(err, "failed to setup gcp provider")
	}

	//create namespace resource
	createdNamespace, err := kubernetescorev1.NewNamespace(ctx,
		locals.Namespace,
		&kubernetescorev1.NamespaceArgs{
			Metadata: metav1.ObjectMetaPtrInput(&metav1.ObjectMetaArgs{
				Name:   pulumi.String(locals.Namespace),
				Labels: pulumi.ToStringMap(locals.Labels),
			}),
		}, pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "5s", Update: "5s", Delete: "5s"}),
		pulumi.Provider(kubernetesProvider))
	if err != nil {
		return errors.Wrapf(err, "failed to create %s namespace", locals.Namespace)
	}

	//export name of the namespace
	ctx.Export(outputs.Namespace, createdNamespace.Metadata.Name())

	//install the openfga helm-chart
	if err := helmChart(ctx, locals, createdNamespace); err != nil {
		return errors.Wrap(err, "failed to create helm-chart resources")
	}

	//create istio-ingress resources if ingress is enabled.
	if locals.OpenfgaKubernetes.Spec.Ingress.IsEnabled {
		if err := ingress(ctx, locals, createdNamespace, kubernetesProvider, locals.Labels); err != nil {
			return errors.Wrap(err, "failed to create ingress resources")
		}
	}

	return nil
}
