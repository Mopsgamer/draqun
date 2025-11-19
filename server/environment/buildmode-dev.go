//go:build !prod

package environment

const BuildEnvironment BuildMode = BuildModeDevelopment
const BuildEnvironmentName string = "Development"
