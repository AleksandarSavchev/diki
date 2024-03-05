// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package v1r11

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	kubeutils "github.com/gardener/diki/pkg/kubernetes/utils"
	"github.com/gardener/diki/pkg/rule"
	sharedv1r11 "github.com/gardener/diki/pkg/shared/ruleset/disak8sstig/v1r11"
)

var _ rule.Rule = &Rule242400{}

type Rule242400 struct {
	Client          client.Client
	Namespace       string
	DeploymentNames []string
	ContainerNames  []string
}

func (r *Rule242400) ID() string {
	return sharedv1r11.ID242400
}

func (r *Rule242400) Name() string {
	return "The Kubernetes API server must have Alpha APIs disabled (MEDIUM 242400)"
}

func (r *Rule242400) Run(ctx context.Context) (rule.RuleResult, error) {
	const option = "feature-gates.AllAlpha"
	var (
		checkResults    []rule.CheckResult
		deploymentNames = []string{"kube-apiserver", "kube-controller-manager", "kube-scheduler"}
		containerNames  = []string{"kube-apiserver", "kube-controller-manager", "kube-scheduler"}
	)

	if r.DeploymentNames != nil {
		deploymentNames = r.DeploymentNames
	}

	if r.ContainerNames != nil {
		containerNames = r.ContainerNames
	}

	for idx, deploymentName := range deploymentNames {
		target := rule.NewTarget("name", deploymentName, "namespace", r.Namespace, "kind", "deployment")

		fgOptions, err := kubeutils.GetCommandOptionFromDeployment(ctx, r.Client, deploymentName, containerNames[idx], r.Namespace, "feature-gates")
		if err != nil {
			checkResults = append(checkResults, rule.ErroredCheckResult(err.Error(), target))
			continue
		}

		allAlphaOptions := kubeutils.FindInnerValue(fgOptions, "AllAlpha")

		// featureGates.AllAlpha defaults to false. ref https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/
		switch {
		case len(allAlphaOptions) == 0:
			checkResults = append(checkResults, rule.PassedCheckResult(fmt.Sprintf("Option %s not set.", option), target))
		case len(allAlphaOptions) > 1:
			checkResults = append(checkResults, rule.WarningCheckResult(fmt.Sprintf("Option %s set more than once in container command.", option), target))
		case allAlphaOptions[0] == "true":
			checkResults = append(checkResults, rule.FailedCheckResult(fmt.Sprintf("Option %s set to not allowed value.", option), target))
		case allAlphaOptions[0] == "false":
			checkResults = append(checkResults, rule.PassedCheckResult(fmt.Sprintf("Option %s set to allowed value.", option), target))
		default:
			checkResults = append(checkResults, rule.WarningCheckResult(fmt.Sprintf("Option %s set to neither 'true' nor 'false'.", option), target))
		}
	}
	return rule.RuleResult{
		RuleID:       r.ID(),
		RuleName:     r.Name(),
		CheckResults: checkResults,
	}, nil
}
