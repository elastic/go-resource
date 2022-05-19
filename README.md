# resource - standalone configuration management

`resource` is a Go package for declarative [configuration
management](https://en.wikipedia.org/wiki/Software_configuration_management)
in a system, in the line of
[Puppet](https://en.wikipedia.org/wiki/Puppet_%28software%29) or [Chef](https://en.wikipedia.org/wiki/Progress_Chef), but embedded in Go binaries.

It is intended to be idempotent, and stateless, two executions of the same
code in the same system should produce the same result.

Some use cases:

* Define test or development scenarios in a declarative way.
* Configure systems with a single static binary, in a [cloudinit](https://wiki.archlinux.org/title/Cloud-init) fashion.
* *[TBD] Configure infrastructure from serverless functions.*

## resource framework

The framework for `resource` is based on the following concepts:

* Facts: information obtained from the environment to customize the execution and not dependant
  on the defined resources (OS information, environment variables...).
* Resources: the actual resources to manage. Their semantics would be: Get/Create/Update.
* Providers: implementations of the resources, they can contain configuration. There can be multiple
  instances of the same provider, resources should be able to select which one to use, with a
  default one.
* Manager: processes all defined resources, generates a plan and executes it.

Some extras that are being considered or in development:

* Conditions: To run resources depending on facts.
* Dependencies: To control the order of execution of resources.
* Migrations: allow to version configurations, and implement migration
  (and rollback?) processes that cannot be managed by resources themselves.
* Modules: Parameterizable collections of resources.

## Getting started

You can start using this package by importing it.
```golang
import "github.com/elastic/go-resource"
```

Find here an example that creates some files for a docker compose
scenario that starts the Elastic Stack:

```golang
package main

import (
	"embed"
        "log"

        "github.com/elastic/go-resource"
)

//go:embed _static
var static embed.FS

var (
        // Define a source of files from an embedded file system
	// You can include additional functions for templates.
        templateFuncs = template.FuncMap{
                "semverLessThan": semverLessThan,
        }
        staticSource     = resource.NewSourceFS(static).WithTemplateFuncs(templateFuncs)

        // Define the resources.
        stackResources = []resource.Resource{
		// Files can be defined as static files, or as templates.
                &resource.File{
                        Provider: "stack-file",
                        Path:     "Dockerfile.package-registry",
                        Content:  staticSource.Template("_static/Dockerfile.package-registry.tmpl"),
                },
                &resource.File{
                        Provider: "stack-file",
                        Path:     "docker-compose.yml",
                        Content:  staticSource.File("_static/docker-compose-stack.yml"),
                },
                &resource.File{
                        Provider: "stack-file",
                        Path:     "elasticsearch.yml",
                        Content:  staticSource.Template("_static/elasticsearch.yml.tmpl"),
                },
                &resource.File{
                        Provider: "stack-file",
                        Path:     "kibana.yml",
                        Content:  staticSource.Template("_static/kibana.yml.tmpl"),
                },
                &resource.File{
                        Provider: "stack-file",
                        Path:     "package-registry.yml",
                        Content:  staticSource.File("_static/package-registry.yml"),
                },
        }
)

func main() {
        // Instantiate a new manager.
        manager := resource.NewManager()

        // Install some facts in the manager. These facts can be
        // used by template files or other resources.
        manager.AddFacter(resource.StaticFacter{
                "registry_base_image":   packageRegistryBaseImage,
                "elasticsearch_version": stackVersion,
                "kibana_version":        stackVersion,
        })

        // Configure a file provider to decide the prefix path where
        // files will be installed.
        manager.RegisterProvider("stack-file", &resource.FileProvider{
                Prefix: stackDir,
        })

        // Apply the defined resources.
        results, err := manager.Apply(stackResources)

        // If there are errors, they can be individually inspected in the
        // returned results.
        if err != nil {
                for _, result := range results {
                        if err := result.Err(); err != nil {
                                log.Println(err)
                        }
                }
                log.Println(err)
        }
}

```

You can find this complete example and others in *TBD*.

## Space, Time

This project started during an ON Week, a time we give each other in Elastic to
explore ideas or learn new things, in alignment with our [Source Code](https://www.elastic.co/about/our-source-code).
