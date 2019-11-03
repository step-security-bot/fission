/*
Copyright 2019 The Fission Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package httptrigger

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	fv1 "github.com/fission/fission/pkg/apis/fission.io/v1"
	"github.com/fission/fission/pkg/controller/client"
	"github.com/fission/fission/pkg/fission-cli/cliwrapper/cli"
	"github.com/fission/fission/pkg/fission-cli/cmd"
)

type GetSubCommand struct {
	client    *client.Client
	trigger   string
	namespace string
}

func Get(flags cli.Input) error {
	opts := GetSubCommand{
		client: cmd.GetServer(flags),
	}
	return opts.do(flags)
}

func (opts *GetSubCommand) do(flags cli.Input) error {
	err := opts.complete(flags)
	if err != nil {
		return err
	}
	return opts.run(flags)
}

// complete creates a environment objects and populates it with default value and CLI inputs.
func (opts *GetSubCommand) complete(flags cli.Input) error {
	opts.trigger = flags.String("name")
	opts.namespace = flags.String("fnNamespace")

	if len(opts.trigger) <= 0 {
		return errors.New("need a trigger name, use --name")
	}
	return nil
}

func (opts *GetSubCommand) run(flags cli.Input) error {
	m := &metav1.ObjectMeta{
		Name:      opts.trigger,
		Namespace: opts.namespace,
	}
	ht, err := opts.client.HTTPTriggerGet(m)
	if err != nil {
		return errors.Wrap(err, "error getting http trigger")
	}

	printHtSummary([]fv1.HTTPTrigger{*ht})

	return nil
}

func printHtSummary(triggers []fv1.HTTPTrigger) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n", "NAME", "METHOD", "URL", "FUNCTION(s)", "INGRESS", "HOST", "PATH", "TLS", "ANNOTATIONS")
	for _, trigger := range triggers {
		function := ""
		if trigger.Spec.FunctionReference.Type == fv1.FunctionReferenceTypeFunctionName {
			function = trigger.Spec.FunctionReference.Name
		} else {
			for k, v := range trigger.Spec.FunctionReference.FunctionWeights {
				function += fmt.Sprintf("%s:%v ", k, v)
			}
		}

		host := trigger.Spec.Host
		if len(trigger.Spec.IngressConfig.Host) > 0 {
			host = trigger.Spec.IngressConfig.Host
		}
		path := trigger.Spec.RelativeURL
		if len(trigger.Spec.IngressConfig.Path) > 0 {
			path = trigger.Spec.IngressConfig.Path
		}

		var msg []string
		for k, v := range trigger.Spec.IngressConfig.Annotations {
			msg = append(msg, fmt.Sprintf("%v: %v", k, v))
		}
		ann := strings.Join(msg, ", ")

		fmt.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\t%v\n",
			trigger.Metadata.Name, trigger.Spec.Method, trigger.Spec.RelativeURL, function, trigger.Spec.CreateIngress, host, path, trigger.Spec.IngressConfig.TLS, ann)
	}
	w.Flush()
}