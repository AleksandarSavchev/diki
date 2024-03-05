// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1r11_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/gardener/diki/pkg/provider/virtualgarden/ruleset/disak8sstig/v1r11"
	"github.com/gardener/diki/pkg/rule"
)

var _ = Describe("#242400", func() {
	var (
		fakeClient      client.Client
		ctx             = context.TODO()
		namespace       = "foo"
		plainDeployment *appsv1.Deployment
		kapiDeployment  *appsv1.Deployment
		kcmDeployment   *appsv1.Deployment
		ksDeployment    *appsv1.Deployment
		target          = rule.NewTarget("namespace", namespace, "kind", "deployment")
	)

	BeforeEach(func() {
		fakeClient = fakeclient.NewClientBuilder().Build()

		plainDeployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Command: []string{},
								Args:    []string{},
							},
						},
					},
				},
			},
		}

		ksDeployment = &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kube-apiserver",
				Namespace: namespace,
			},
			Spec: appsv1.DeploymentSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:    "kube-apiserver",
								Command: []string{},
								Args:    []string{},
							},
						},
					},
				},
			},
		}
	})

	It("should error when deployment are not found", func() {
		r := &v1r11.Rule242400{Client: fakeClient, Namespace: namespace}

		ruleResult, err := r.Run(ctx)
		Expect(err).ToNot(HaveOccurred())

		Expect(ruleResult.CheckResults).To(Equal([]rule.CheckResult{
			{
				Status:  rule.Errored,
				Message: "deployments.apps \"kube-apiserver\" not found",
				Target:  target.With("name", "kube-apiserver"),
			},
			{
				Status:  rule.Errored,
				Message: "deployments.apps \"kube-controller-manager\" not found",
				Target:  target.With("name", "kube-controller-manager"),
			},
			{
				Status:  rule.Errored,
				Message: "deployments.apps \"kube-scheduler\" not found",
				Target:  target.With("name", "kube-scheduler"),
			},
		},
		))
	})

	It("should return correct checkResults", func() {
		kapiDeployment = plainDeployment.DeepCopy()
		kapiDeployment.Name = "kube-apiserver"
		kapiDeployment.Spec.Template.Spec.Containers[0].Name = "kube-apiserver"
		kapiDeployment.Spec.Template.Spec.Containers[0].Command = []string{"--flag1=value1", "--flag2=value2"}

		kcmDeployment = plainDeployment.DeepCopy()
		kcmDeployment.Name = "kube-controller-manager"
		kcmDeployment.Spec.Template.Spec.Containers[0].Name = "kube-controller-manager"
		kcmDeployment.Spec.Template.Spec.Containers[0].Command = []string{"--feature-gates=AllAlpha=false,AllBeta=true"}

		ksDeployment = plainDeployment.DeepCopy()
		ksDeployment.Name = "kube-scheduler"
		ksDeployment.Spec.Template.Spec.Containers[0].Name = "kube-scheduler"
		ksDeployment.Spec.Template.Spec.Containers[0].Command = []string{"--feature-gates=AllAlpha=true"}

		Expect(fakeClient.Create(ctx, kapiDeployment)).To(Succeed())
		Expect(fakeClient.Create(ctx, kcmDeployment)).To(Succeed())
		Expect(fakeClient.Create(ctx, ksDeployment)).To(Succeed())

		r := &v1r11.Rule242400{Client: fakeClient, Namespace: namespace}
		ruleResult, err := r.Run(ctx)

		expectedCheckResults := []rule.CheckResult{
			rule.PassedCheckResult("Option feature-gates.AllAlpha not set.", rule.NewTarget("kind", "deployment", "name", "kube-apiserver", "namespace", "foo")),
			rule.PassedCheckResult("Option feature-gates.AllAlpha set to allowed value.", rule.NewTarget("kind", "deployment", "name", "kube-controller-manager", "namespace", "foo")),
			rule.FailedCheckResult("Option feature-gates.AllAlpha set to not allowed value.", rule.NewTarget("kind", "deployment", "name", "kube-scheduler", "namespace", "foo")),
		}

		Expect(err).To(BeNil())
		Expect(ruleResult.CheckResults).To(Equal(expectedCheckResults))
	})

	It("should return correct warning when options are not set properly", func() {
		kapiDeployment = plainDeployment.DeepCopy()
		kapiDeployment.Name = "kube-apiserver"
		kapiDeployment.Spec.Template.Spec.Containers[0].Name = "kube-apiserver"
		kapiDeployment.Spec.Template.Spec.Containers[0].Command = []string{"--feature-gates=AllAlpha=true", "--feature-gates=AllAlpha=false"}

		kcmDeployment = plainDeployment.DeepCopy()
		kcmDeployment.Name = "kube-controller-manager"
		kcmDeployment.Spec.Template.Spec.Containers[0].Name = "kube-controller-manager"
		kcmDeployment.Spec.Template.Spec.Containers[0].Command = []string{"--feature-gates=AllAlpha=not-false,AllBeta=true"}

		ksDeployment = plainDeployment.DeepCopy()
		ksDeployment.Name = "kube-scheduler"
		ksDeployment.Spec.Template.Spec.Containers[0].Name = "kube-scheduler"
		ksDeployment.Spec.Template.Spec.Containers[0].Command = []string{"--feature-gates=AllAlpha=not-true", "--feature-gates=AllAlpha=false"}

		Expect(fakeClient.Create(ctx, kapiDeployment)).To(Succeed())
		Expect(fakeClient.Create(ctx, kcmDeployment)).To(Succeed())
		Expect(fakeClient.Create(ctx, ksDeployment)).To(Succeed())

		r := &v1r11.Rule242400{Client: fakeClient, Namespace: namespace}
		ruleResult, err := r.Run(ctx)

		expectedCheckResults := []rule.CheckResult{
			rule.WarningCheckResult("Option feature-gates.AllAlpha set more than once in container command.", rule.NewTarget("kind", "deployment", "name", "kube-apiserver", "namespace", "foo")),
			rule.WarningCheckResult("Option feature-gates.AllAlpha set to neither 'true' nor 'false'.", rule.NewTarget("kind", "deployment", "name", "kube-controller-manager", "namespace", "foo")),
			rule.WarningCheckResult("Option feature-gates.AllAlpha set more than once in container command.", rule.NewTarget("kind", "deployment", "name", "kube-scheduler", "namespace", "foo")),
		}

		Expect(err).To(BeNil())
		Expect(ruleResult.CheckResults).To(Equal(expectedCheckResults))
	})

	It("should return correc checkResults only for selected deployments", func() {
		kapiDeployment = plainDeployment.DeepCopy()
		kapiDeployment.Name = "kube-apiserver"
		kapiDeployment.Spec.Template.Spec.Containers[0].Name = "kube-apiserver"
		kapiDeployment.Spec.Template.Spec.Containers[0].Command = []string{"--flag1=value1", "--flag2=value2"}

		kcmDeployment = plainDeployment.DeepCopy()
		kcmDeployment.Name = "kube-controller-manager"
		kcmDeployment.Spec.Template.Spec.Containers[0].Name = "kube-controller-manager"
		kcmDeployment.Spec.Template.Spec.Containers[0].Command = []string{"--feature-gates=AllAlpha=false,AllBeta=true"}

		ksDeployment = plainDeployment.DeepCopy()
		ksDeployment.Name = "kube-scheduler"
		ksDeployment.Spec.Template.Spec.Containers[0].Name = "kube-scheduler"
		ksDeployment.Spec.Template.Spec.Containers[0].Command = []string{"--feature-gates=AllAlpha=true"}

		Expect(fakeClient.Create(ctx, kapiDeployment)).To(Succeed())
		Expect(fakeClient.Create(ctx, kcmDeployment)).To(Succeed())
		Expect(fakeClient.Create(ctx, ksDeployment)).To(Succeed())

		r := &v1r11.Rule242400{
			Client:          fakeClient,
			Namespace:       namespace,
			DeploymentNames: []string{"kube-apiserver"},
			ContainerNames:  []string{"kube-apiserver"},
		}
		ruleResult, err := r.Run(ctx)

		expectedCheckResults := []rule.CheckResult{
			rule.PassedCheckResult("Option feature-gates.AllAlpha not set.", rule.NewTarget("kind", "deployment", "name", "kube-apiserver", "namespace", "foo")),
		}

		Expect(err).To(BeNil())
		Expect(ruleResult.CheckResults).To(Equal(expectedCheckResults))
	})
})
