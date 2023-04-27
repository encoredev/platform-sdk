// Package platform is the SDK for communicating with [Encore] hosted services
//
// The [Encore] Platform SDK provides APIs for Encore applications to use at
// runtime to communicate with Encore hosted services. It is used by the
// [Encore Runtime] as a provide SDK for communicating with the Encore Platform,
// the same way the [Encore Runtime] uses the AWS or GCP SDKs under the hood.
//
// It is not intended to be used by applications directly, and therefore it's
// API is not considered to be stable.
//
// # Overview of Packages
//
//   - platform - The main SDK package, contains shared types used
//   - encorecloud - The SDK for communicating with the Encore Cloud specific services
//
// [Encore]: https://encore.dev
// [Encore Runtime]: https://github.com/encoredev/encore/tree/main/runtime
package platform
