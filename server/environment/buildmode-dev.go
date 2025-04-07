//go:build !prod && !test

package environment

const BuildModeValue BuildMode = BuildModeDevelopment
const BuildModeName string = "development"
