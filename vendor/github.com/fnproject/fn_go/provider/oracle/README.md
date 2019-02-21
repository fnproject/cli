# Oracle Provider

This provides Oracle Cloud Infrastructure (OCI) signing support for the FN golang SDK 

Configuration:

The provider can read most of it's settings from [OCI configuration file](https://docs.us-phoenix-1.oraclecloud.com/Content/API/Concepts/sdkconfig.htm): (~/.oci/config)

|  Key               | Example      |  Required | Read from ~/.oci/config | Description |
| -------------------|  ----------- |  -----    | ----- |  ---- |  
| `api-url` | https://api.faas.us-ashburn-1.oraclecloud.com/ | Yes | No | The API endpoint to contact for accessing the service API |
| `call-url` | https://r.faas.us-ashburn-1.oraclecloud.com/  | Yes (If calls are being made)  | No | The call endpoint  base URL for calling functions |
| `oracle.compartment-id` | ocid1.compartment.oc1..aaaaaaaajvunnz..... | Yes | No | The compartment OCID for the functions tenancy - this corresponds to where you want functions objects to exist in OCI |
| `oracle.tenancy-id` | ocid1.tenancy.oc1..aaaaaaaai4w6iipzc73k3s2.... | No | The tenancy of the user accessing the service |
| `oracle.user-id` | ocid1.user.oc1..aaaaaaaay2df7zq4lgv7.... | No | Yes | The OCID of the user accessing the API |
| `oracle.fingerprint`|  30:c1:f8:98:38:be:bb... | No | Yes | The RSA key fingerprint of the key being used |
| `oracle.key-file` | /home/myuser/.oci/private_key.pem | No | Yes (`key_file`) | The private key for the registered API key |
| `oracle.pass-phrase`|  | No | Yes | (`pass_phrase` ) | The passphrase for the private key file - if unspecified this will be requested from the configured passphrase source |
| `oracle.profile` | | No |  No | Defaults to `DEFAULT`  - the OCI Configuration profile to use for reading OCI information |
| `oracle.disable-certs` |`true`| No | No | Ignore SSL host name checks when contacting the server (should only be used for diagnosis and testing) |
