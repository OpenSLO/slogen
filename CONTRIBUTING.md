## adding support for a new source type


###### implement a new package under `libs` for that source

The package should expose function `IsSource` that given a parsed SLO object returns true if the source is supported by the package. 
An example for sumologic source detection is [here](https://github.com/OpenSLO/slogen/blob/sumo-agaurav/libs/sumologic/tf.go#L78-L92)

The package should also expose a function `GiveTerraform` that returns the terraform content to be added for that source when provided with the parsed slo configs.
The above function can then be called in [`libs/gen.go`](https://github.com/OpenSLO/slogen/blob/sumo-agaurav/libs/gen.go#L111) to create the corresponding terraform files.


###### add the required terraform providers for that source in [libs/templates/terraform//main.tf.gotf](libs/templates/terraform/main.tf.gotf)

