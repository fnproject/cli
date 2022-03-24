// Copyright (c) 2016, 2018, 2021, Oracle and/or its affiliates.  All rights reserved.
// This software is dual-licensed to you under the Universal Permissive License (UPL) 1.0 as shown at https://oss.oracle.com/licenses/upl or Apache License 2.0 as shown at http://www.apache.org/licenses/LICENSE-2.0. You may choose either license.
// Code generated. DO NOT EDIT.

// Container Images API
//
// API covering the Registry (https://docs.cloud.oracle.com/iaas/Content/Registry/Concepts/registryoverview.htm) services.
// Use this API to manage resources such as container images and repositories.
//

package artifacts

import (
	"github.com/oracle/oci-go-sdk/v48/common"
)

// ContainerImageSignature Container image signature metadata.
type ContainerImageSignature struct {

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the compartment in which the container repository exists.
	CompartmentId *string `mandatory:"true" json:"compartmentId"`

	// The id of the user or principal that created the resource.
	CreatedBy *string `mandatory:"true" json:"createdBy"`

	// The last 10 characters of the kmsKeyId, the last 10 characters of the kmsKeyVersionId, the signingAlgorithm, and the last 10 characters of the signatureId.
	// Example: `wrmz22sixa::qdwyc2ptun::SHA_256_RSA_PKCS_PSS::2vwmobasva`
	DisplayName *string `mandatory:"true" json:"displayName"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the container image signature.
	// Example: `ocid1.containerimagesignature.oc1..exampleuniqueID`
	Id *string `mandatory:"true" json:"id"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the container image.
	// Example: `ocid1.containerimage.oc1..exampleuniqueID`
	ImageId *string `mandatory:"true" json:"imageId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the kmsKeyId used to sign the container image.
	// Example: `ocid1.key.oc1..exampleuniqueID`
	KmsKeyId *string `mandatory:"true" json:"kmsKeyId"`

	// The OCID (https://docs.cloud.oracle.com/Content/General/Concepts/identifiers.htm) of the kmsKeyVersionId used to sign the container image.
	// Example: `ocid1.keyversion.oc1..exampleuniqueID`
	KmsKeyVersionId *string `mandatory:"true" json:"kmsKeyVersionId"`

	// The base64 encoded signature payload that was signed.
	Message *string `mandatory:"true" json:"message"`

	// The signature of the message field using the kmsKeyId, the kmsKeyVersionId, and the signingAlgorithm.
	Signature *string `mandatory:"true" json:"signature"`

	// The algorithm to be used for signing. These are the only supported signing algorithms for container images.
	SigningAlgorithm ContainerImageSignatureSigningAlgorithmEnum `mandatory:"true" json:"signingAlgorithm"`

	// An RFC 3339 timestamp indicating when the image was created.
	TimeCreated *common.SDKTime `mandatory:"true" json:"timeCreated"`
}

func (m ContainerImageSignature) String() string {
	return common.PointerString(m)
}

// ContainerImageSignatureSigningAlgorithmEnum Enum with underlying type: string
type ContainerImageSignatureSigningAlgorithmEnum string

// Set of constants representing the allowable values for ContainerImageSignatureSigningAlgorithmEnum
const (
	ContainerImageSignatureSigningAlgorithm224RsaPkcsPss ContainerImageSignatureSigningAlgorithmEnum = "SHA_224_RSA_PKCS_PSS"
	ContainerImageSignatureSigningAlgorithm256RsaPkcsPss ContainerImageSignatureSigningAlgorithmEnum = "SHA_256_RSA_PKCS_PSS"
	ContainerImageSignatureSigningAlgorithm384RsaPkcsPss ContainerImageSignatureSigningAlgorithmEnum = "SHA_384_RSA_PKCS_PSS"
	ContainerImageSignatureSigningAlgorithm512RsaPkcsPss ContainerImageSignatureSigningAlgorithmEnum = "SHA_512_RSA_PKCS_PSS"
)

var mappingContainerImageSignatureSigningAlgorithm = map[string]ContainerImageSignatureSigningAlgorithmEnum{
	"SHA_224_RSA_PKCS_PSS": ContainerImageSignatureSigningAlgorithm224RsaPkcsPss,
	"SHA_256_RSA_PKCS_PSS": ContainerImageSignatureSigningAlgorithm256RsaPkcsPss,
	"SHA_384_RSA_PKCS_PSS": ContainerImageSignatureSigningAlgorithm384RsaPkcsPss,
	"SHA_512_RSA_PKCS_PSS": ContainerImageSignatureSigningAlgorithm512RsaPkcsPss,
}

// GetContainerImageSignatureSigningAlgorithmEnumValues Enumerates the set of values for ContainerImageSignatureSigningAlgorithmEnum
func GetContainerImageSignatureSigningAlgorithmEnumValues() []ContainerImageSignatureSigningAlgorithmEnum {
	values := make([]ContainerImageSignatureSigningAlgorithmEnum, 0)
	for _, v := range mappingContainerImageSignatureSigningAlgorithm {
		values = append(values, v)
	}
	return values
}
