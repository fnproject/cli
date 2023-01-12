/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package commands

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fnproject/fn_go/provider/oracle"

	client "github.com/fnproject/cli/client"
	common "github.com/fnproject/cli/common"
	apps "github.com/fnproject/cli/objects/app"
	function "github.com/fnproject/cli/objects/fn"
	trigger "github.com/fnproject/cli/objects/trigger"
	v2Client "github.com/fnproject/fn_go/clientv2"
	models "github.com/fnproject/fn_go/modelsv2"
	"github.com/oracle/oci-go-sdk/v48/artifacts"
	ociCommon "github.com/oracle/oci-go-sdk/v48/common"
	"github.com/oracle/oci-go-sdk/v48/keymanagement"
	"github.com/urfave/cli"
)

// Message defines the struct of container image signature payload
type Message struct {
	Description      string `mandatory:"true" json:"description"`
	ImageDigest      string `mandatory:"true" json:"imageDigest"`
	KmsKeyId         string `mandatory:"true" json:"kmsKeyId"`
	KmsKeyVersionId  string `mandatory:"true" json:"kmsKeyVersionId"`
	Metadata         string `mandatory:"true" json:"metadata"`
	Region           string `mandatory:"true" json:"region"`
	RepositoryName   string `mandatory:"true" json:"repositoryName"`
	SigningAlgorithm string `mandatory:"true" json:"signingAlgorithm"`
}

var RegionsWithOldKMSEndpoints = map[ociCommon.Region]struct{}{
	ociCommon.RegionSEA:           {},
	ociCommon.RegionPHX:           {},
	ociCommon.RegionIAD:           {},
	ociCommon.RegionFRA:           {},
	ociCommon.RegionLHR:           {},
	ociCommon.RegionCAToronto1:    {},
	ociCommon.RegionAPSeoul1:      {},
	ociCommon.RegionAPTokyo1:      {},
	ociCommon.RegionAPMumbai1:     {},
	ociCommon.RegionEUZurich1:     {},
	ociCommon.RegionSASaopaulo1:   {},
	ociCommon.RegionAPSydney1:     {},
	ociCommon.RegionMEJeddah1:     {},
	ociCommon.RegionEUAmsterdam1:  {},
	ociCommon.RegionAPMelbourne1:  {},
	ociCommon.RegionAPOsaka1:      {},
	ociCommon.RegionCAMontreal1:   {},
	ociCommon.RegionUSLangley1:    {},
	ociCommon.RegionUSLuke1:       {},
	ociCommon.RegionUSGovAshburn1: {},
	ociCommon.RegionUSGovChicago1: {},
	ociCommon.RegionUSGovPhoenix1: {},
	ociCommon.RegionUKGovLondon1:  {},
}

// DeployCommand returns deploy cli.command
func DeployCommand() cli.Command {
	cmd := deploycmd{}
	var flags []cli.Flag
	flags = append(flags, cmd.flags()...)
	return cli.Command{
		Name:    "deploy",
		Usage:   "\tDeploys a function to the functions server (bumps, build, pushes and updates functions and/or triggers).",
		Aliases: []string{"dp"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			cmd.clientV2 = provider.APIClientv2()
			return nil
		},
		Category:    "DEVELOPMENT COMMANDS",
		Description: "This command deploys one or all (--all) functions to the function server.",
		ArgsUsage:   "[function-subdirectory]",
		Flags:       flags,
		Action:      cmd.deploy,
	}
}

type deploycmd struct {
	clientV2 *v2Client.Fn

	appName   string
	createApp bool
	wd        string
	local     bool
	noCache   bool
	registry  string
	all       bool
	noBump    bool
}

func (p *deploycmd) flags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "app",
			Usage:       "App name to deploy to",
			Destination: &p.appName,
		},
		cli.BoolFlag{
			Name:        "create-app",
			Usage:       "Enable automatic creation of app if it doesn't exist during deploy",
			Destination: &p.createApp,
		},
		cli.BoolFlag{
			Name:        "verbose, v",
			Usage:       "Verbose mode",
			Destination: &common.CommandVerbose,
		},
		cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "Don't use Docker cache for the build",
			Destination: &p.noCache,
		},
		cli.BoolFlag{
			Name:        "local, skip-push", // todo: deprecate skip-push
			Usage:       "Do not push Docker built images onto Docker Hub - useful for local development.",
			Destination: &p.local,
		},
		cli.StringFlag{
			Name:        "registry",
			Usage:       "Set the Docker owner for images and optionally the registry. This will be prefixed to your function name for pushing to Docker registries.\r  eg: `--registry username` will set your Docker Hub owner. `--registry registry.hub.docker.com/username` will set the registry and owner. ",
			Destination: &p.registry,
		},
		cli.BoolFlag{
			Name:        "all",
			Usage:       "If in root directory containing `app.yaml`, this will deploy all functions",
			Destination: &p.all,
		},
		cli.BoolFlag{
			Name:        "no-bump",
			Usage:       "Do not bump the version, assuming external version management",
			Destination: &p.noBump,
		},
		cli.StringSliceFlag{
			Name:  "build-arg",
			Usage: "Set build time variables",
		},
		cli.StringFlag{
			Name:  "working-dir,w",
			Usage: "Specify the working directory to deploy a function, must be the full path.",
		},
	}
}

// deploy deploys a function or a set of functions for an app
// By default this will deploy a single function, either the function in the current directory
// or if an arg is passed in, a function in the path representing that arg, relative to the
// current working directory.
//
// If user passes in --all flag, it will deploy all functions in an app. An app must have an `app.yaml`
// file in it's root directory. The functions will be deployed based on the directory structure
// on the file system (can be overridden using the `path` arg in each `func.yaml`. The index/root function
// is the one that lives in the same directory as the app.yaml.
func (p *deploycmd) deploy(c *cli.Context) error {
	appName := ""
	dir := common.GetDir(c)

	appf, err := common.LoadAppfile(dir)
	if err != nil {
		if _, ok := err.(*common.NotFoundError); ok {
			if p.all {
				return err
			}
			// otherwise, it's ok
		} else {
			return err
		}
	} else {
		appName = appf.Name
	}
	if p.appName != "" {
		// flag overrides all
		appName = p.appName
	}

	if appName == "" {
		return errors.New("App name must be provided, try `--app APP_NAME`")
	}

	// appfApp is used to create/update app, with app file additions if provided
	appfApp := models.App{
		Name: appName,
	}
	if appf != nil {
		// set other fields from app file
		appfApp.Config = appf.Config
		appfApp.Annotations = appf.Annotations
		if appf.SyslogURL != "" {
			// TODO consistent with some other fields (config), unsetting in app.yaml doesn't unset on server. undecided policy for all fields
			appfApp.SyslogURL = &appf.SyslogURL
		}
	}

	// find and create/update app if required
	app, err := apps.GetAppByName(p.clientV2, appName)
	if _, ok := err.(apps.NameNotFoundError); ok && p.createApp {
		app, err = apps.CreateApp(p.clientV2, &appfApp)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else if appf != nil {
		// app exists, but we need to update it if we have an app file
		app, err = apps.PutApp(p.clientV2, app.ID, &appfApp)
		if err != nil {
			return fmt.Errorf("Failed to update app config: %v", err)
		}
	}

	if app == nil {
		panic("app should not be nil here") // tests should catch... better than panic later
	}

	// deploy functions
	if p.all {
		return p.deployAll(c, app)
	}
	return p.deploySingle(c, app)
}

// deploySingle deploys a single function, either the current directory or if in the context
// of an app and user provides relative path as the first arg, it will deploy that function.
func (p *deploycmd) deploySingle(c *cli.Context, app *models.App) error {
	var dir string
	wd := common.GetWd()

	if c.String("working-dir") != "" {
		dir = c.String("working-dir")
	} else {
		// if we're in the context of an app, first arg is path to the function
		path := c.Args().First()
		if path != "" {
			fmt.Printf("Deploying function at: ./%s\n", path)
		}
		dir = filepath.Join(wd, path)
	}

	err := os.Chdir(dir)
	if err != nil {
		return err
	}
	defer os.Chdir(wd)

	fpath, ff, err := common.FindAndParseFuncFileV20180708(dir)
	if err != nil {
		return err
	}
	return p.deployFuncV20180708(c, app, fpath, ff)
}

// deployAll deploys all functions in an app.
func (p *deploycmd) deployAll(c *cli.Context, app *models.App) error {
	var dir string
	wd := common.GetWd()

	if c.String("working-dir") != "" {
		dir = c.String("working-dir")
	} else {
		// if we're in the context of an app, first arg is path to the function
		path := c.Args().First()
		if path != "" {
			fmt.Printf("Deploying function at: ./%s\n", path)
		}
		dir = filepath.Join(wd, path)
	}

	var funcFound bool
	err := common.WalkFuncsV20180708(dir, func(path string, ff *common.FuncFileV20180708, err error) error {
		if err != nil { // probably some issue with funcfile parsing, can decide to handle this differently if we'd like
			return err
		}
		dir := filepath.Dir(path)
		if dir != wd {
			// change dirs
			err = os.Chdir(dir)
			if err != nil {
				return err
			}
		}
		p2 := strings.TrimPrefix(dir, wd)
		if ff.Name == "" {
			ff.Name = strings.Replace(p2, "/", "-", -1)
			if strings.HasPrefix(ff.Name, "-") {
				ff.Name = ff.Name[1:]
			}
		}

		err = p.deployFuncV20180708(c, app, path, ff)
		if err != nil {
			return fmt.Errorf("deploy error on %s: %v", path, err)
		}

		now := time.Now()
		os.Chtimes(path, now, now)
		funcFound = true
		return nil
	})
	if err != nil {
		return err
	}

	if !funcFound {
		return errors.New("No functions found to deploy")
	}

	return nil
}

func (p *deploycmd) deployFuncV20180708(c *cli.Context, app *models.App, funcfilePath string, funcfile *common.FuncFileV20180708) error {
	if funcfile.Name == "" {
		funcfile.Name = filepath.Base(filepath.Dir(funcfilePath)) // todo: should probably make a copy of ff before changing it
	}

	oracleProvider, _ := getOracleProvider()
	if oracleProvider != nil && oracleProvider.ImageCompartmentID != "" {
		// If the provider is Oracle and ImageCompartmentID is present, we need to deploy image to the ImageCompartmentID.
		// The repository name should be unique throughout a tenancy. We check if a repository exists in the compartment and create it if it doesn't already exist.
		// If the creation fails, it could be because the repository name aready exists in a different compartment.

		repositoryName, err := getRepositoryName(funcfile)
		if err != nil {
			return err
		}

		artifactsClient, err := artifacts.NewArtifactsClientWithConfigurationProvider(oracleProvider.ConfigurationProvider)
		if err != nil {
			return err
		}
		artifactsClient.SetRegion(getRegion(oracleProvider))

		repositoryExists, err := doesRepositoryExistInCompartment(repositoryName, oracleProvider.ImageCompartmentID, artifactsClient)
		if err != nil {
			return err
		}
		if !repositoryExists {
			err = createContainerRepositoryInCompartment(repositoryName, oracleProvider.ImageCompartmentID, artifactsClient)
			if err != nil {
				return err
			}
		}
	}

	fmt.Printf("Deploying %s to app: %s\n", funcfile.Name, app.Name)
	if !p.noBump {
		funcfile2, err := common.BumpItV20180708(funcfilePath, common.Patch)
		if err != nil {
			return err
		}
		funcfile.Version = funcfile2.Version
		// TODO: this whole funcfile handling needs some love, way too confusing. Only bump makes permanent changes to it.
	}

	buildArgs := c.StringSlice("build-arg")
	_, err := common.BuildFuncV20180708(common.IsVerbose(), funcfilePath, funcfile, buildArgs, p.noCache)
	if err != nil {
		return err
	}

	if !p.local {
		if err := common.PushV20180708(funcfile); err != nil {
			return err
		}
	}

	if err := p.signImage(funcfile); err != nil {
		return err
	}

	return p.updateFunction(c, app.ID, funcfile)
}

func (p *deploycmd) updateFunction(c *cli.Context, appID string, ff *common.FuncFileV20180708) error {
	fmt.Printf("Updating function %s using image %s...\n", ff.Name, ff.ImageNameV20180708())

	fn := &models.Fn{}
	if err := function.WithFuncFileV20180708(ff, fn); err != nil {
		return fmt.Errorf("Error getting function with funcfile: %s", err)
	}

	fnRes, err := function.GetFnByName(p.clientV2, appID, ff.Name)
	if _, ok := err.(function.NameNotFoundError); ok {
		fn.Name = ff.Name
		fn, err = function.CreateFn(p.clientV2, appID, fn)
		if err != nil {
			return err
		}
	} else if err != nil {
		// probably service is down or something...
		return err
	} else {
		fn.ID = fnRes.ID
		err = function.PutFn(p.clientV2, fn.ID, fn)
		if err != nil {
			return err
		}
	}

	if len(ff.Triggers) != 0 {
		for _, t := range ff.Triggers {
			trig := &models.Trigger{
				AppID:  appID,
				FnID:   fn.ID,
				Name:   t.Name,
				Source: t.Source,
				Type:   t.Type,
			}

			trigs, err := trigger.GetTriggerByName(p.clientV2, appID, fn.ID, t.Name)
			if _, ok := err.(trigger.NameNotFoundError); ok {
				err = trigger.CreateTrigger(p.clientV2, trig)
				if err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else {
				trig.ID = trigs.ID
				err = trigger.PutTrigger(p.clientV2, trig)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getOracleProvider() (*oracle.OracleProvider, error) {
	currentProvider, err := client.CurrentProvider()
	if err != nil {
		return nil, err
	}
	if oracleProvider, ok := currentProvider.(*oracle.OracleProvider); ok {
		return oracleProvider, nil
	}
	return nil, nil
}

func (p *deploycmd) signImage(funcfile *common.FuncFileV20180708) error {
	signingDetails := funcfile.SigningDetails
	signatureConfigured, err := isSignatureConfigured(signingDetails)
	if err != nil {
		return err
	}
	if !signatureConfigured {
		return nil
	}
	oracleProvider, _ := getOracleProvider()
	if oracleProvider == nil {
		return nil
	}
	fmt.Printf("Signing image %s using KmsKey %s...\n", funcfile.ImageNameV20180708(), signingDetails.KmsKeyId)
	imageDigest, err := getImageDigest(funcfile)
	if err != nil {
		return err
	}
	fmt.Printf("Image digest is %s\n", imageDigest)
	repositoryName, err := getRepositoryName(funcfile)
	if err != nil {
		return err
	}
	fmt.Printf("Image belongs to repository %s\n", repositoryName)
	artifactsClient, err := artifacts.NewArtifactsClientWithConfigurationProvider(oracleProvider.ConfigurationProvider)
	if err != nil {
		return err
	}
	region := getRegion(oracleProvider)
	artifactsClient.SetRegion(region)
	imageId, compartmentId, err := getImageId(artifactsClient, "", signingDetails.ImageCompartmentId, repositoryName, imageDigest)
	if err != nil {
		return err
	}
	signatureRequired, err := isSignatureRequired(artifactsClient, imageId, signingDetails)
	if err != nil {
		return err
	}
	if !signatureRequired {
		fmt.Printf("Image %s is already signed by %s\n", funcfile.ImageNameV20180708(), signingDetails.KmsKeyId)
		return nil
	}
	message, signature, err := createImageSignature(oracleProvider, region, imageDigest, repositoryName, funcfile.SigningDetails)
	if err != nil {
		return err
	}
	if err = uploadImageSignature(artifactsClient, compartmentId, imageId, message, signature, funcfile.SigningDetails); err == nil {
		fmt.Printf("Successfully signed and uploaded image signature for %s\n", funcfile.ImageNameV20180708())
	}
	return err
}

func isSignatureConfigured(signingDetails common.SigningDetails) (bool, error) {
	configured := signingDetails.SigningAlgorithm != "" && signingDetails.KmsKeyId != "" &&
		signingDetails.ImageCompartmentId != "" && signingDetails.KmsKeyVersionId != ""
	if !configured && (signingDetails.SigningAlgorithm != "" || signingDetails.KmsKeyId != "" ||
		signingDetails.ImageCompartmentId != "" || signingDetails.KmsKeyVersionId != "") {
		return false, fmt.Errorf("signing_details is missing values for [%s] in func.yaml", findMissingValues(signingDetails))
	}
	return configured, nil
}

func getRegion(oracleProvider *oracle.OracleProvider) string {
	// try to derive region from FnApiUrl
	if oracleProvider.FnApiUrl != nil {
		parts := strings.Split(oracleProvider.FnApiUrl.Host, ".")
		if len(parts) >= 4 {
			return parts[1]
		}
	}
	// provider was built after all validations, so it is safe to ignore
	region, _ := oracleProvider.ConfigurationProvider.Region()
	return region
}

func getRepositoryName(ff *common.FuncFileV20180708) (string, error) {
	parts := strings.Split(ff.ImageNameV20180708(), ":")
	if len(parts) != 2 {
		return "", fmt.Errorf("cannot parse image %s", ff.ImageNameV20180708())
	}
	pattern := regexp.MustCompile("(.*)/([^/]*)/(.*)")
	parts = pattern.FindStringSubmatch(parts[0])
	if len(parts) != 4 {
		return "", fmt.Errorf("cannot parse registry for image %s", ff.ImageNameV20180708())
	}
	return parts[3], nil
}

func getImageDigest(ff *common.FuncFileV20180708) (string, error) {
	containerEngineType, err := common.GetContainerEngineType()
	if err != nil {
		return "", err
	}
	fmt.Printf("Fetching image digest for %s\n", ff.ImageNameV20180708())
	parts := strings.Split(ff.ImageNameV20180708(), ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("failed to parse image %s", ff.ImageNameV20180708())
	}
	image, tag := parts[0], parts[1]
	imageDigests, err := exec.Command(containerEngineType, "images", "--digests", image, "--format", "{{.Tag}} {{.Digest}}").Output()
	if err != nil {
		return "", fmt.Errorf("error while listing image digests for %s, %s", ff.ImageNameV20180708(), err)
	}
	cmd := exec.Command("awk", fmt.Sprintf("{if ($1==\"%s\") print $2}", tag))
	cmd.Stdin = bytes.NewBuffer(imageDigests)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error parsing image digest output for %s, %s", ff.ImageNameV20180708(), err)
	}
	imageDigest := strings.ReplaceAll(string(output), "\n", "")
	if imageDigest == "" {
		return "", fmt.Errorf("failed to fetch image digest for %s", ff.ImageNameV20180708())
	}
	return imageDigest, nil
}

func getImageId(client artifacts.ArtifactsClient, page, imageCompartmentId, repositoryName, imageDigest string) (string, string, error) {
	request := artifacts.ListContainerImagesRequest{
		CompartmentId:          ociCommon.String(imageCompartmentId),
		CompartmentIdInSubtree: ociCommon.Bool(true),
		RepositoryName:         ociCommon.String(repositoryName),
	}
	if page != "" {
		request.Page = ociCommon.String(page)
	}
	images, err := client.ListContainerImages(context.Background(), request)
	if err != nil {
		return "", "", fmt.Errorf("failed to lookup image in OCI Registry due to %s", err)
	}
	for _, image := range images.Items {
		if image.Digest != nil && *image.Digest == imageDigest {
			return *image.Id, *image.CompartmentId, nil
		}
	}
	if images.OpcNextPage != nil {
		return getImageId(client, *images.OpcNextPage, imageCompartmentId, repositoryName, imageDigest)
	}
	return "", "", fmt.Errorf("failed to fetch image details for %s from OCI Container Registry", repositoryName)
}

func isSignatureRequired(client artifacts.ArtifactsClient, imageId string, signingDetails common.SigningDetails) (bool, error) {
	algorithmEnum := artifacts.ListContainerImageSignaturesSigningAlgorithmEnum(signingDetails.SigningAlgorithm)
	signatures, err := client.ListContainerImageSignatures(context.Background(), artifacts.ListContainerImageSignaturesRequest{
		CompartmentId:          ociCommon.String(signingDetails.ImageCompartmentId),
		CompartmentIdInSubtree: ociCommon.Bool(true),
		ImageId:                ociCommon.String(imageId),
		KmsKeyId:               ociCommon.String(signingDetails.KmsKeyId),
		KmsKeyVersionId:        ociCommon.String(signingDetails.KmsKeyVersionId),
		SigningAlgorithm:       algorithmEnum,
		Limit:                  ociCommon.Int(1),
	})
	if err != nil {
		return true, err
	}
	return len(signatures.Items) == 0, nil
}

func createImageSignature(provider *oracle.OracleProvider, region string, imageDigest string, repositoryName string, signingDetails common.SigningDetails) (string, string, error) {
	encoded, err := createImageSignatureMessage(region, imageDigest, repositoryName, signingDetails)
	if err != nil {
		return "", "", nil
	}
	algorithm := keymanagement.SignDataDetailsSigningAlgorithmEnum(signingDetails.SigningAlgorithm)
	cryptoEndpoint, err := buildCryptoEndpoint(region, signingDetails.KmsKeyId)
	if err != nil {
		return "", "", fmt.Errorf("failed to build crypto endpoint due to %s", err)
	}
	kmsClient, err := keymanagement.NewKmsCryptoClientWithConfigurationProvider(provider.ConfigurationProvider, cryptoEndpoint)
	if err != nil {
		return "", "", fmt.Errorf("failed to create crypto client due to %s", err)
	}
	signResponse, err := kmsClient.Sign(context.Background(), keymanagement.SignRequest{
		SignDataDetails: keymanagement.SignDataDetails{
			Message:          ociCommon.String(encoded),
			KeyId:            ociCommon.String(signingDetails.KmsKeyId),
			KeyVersionId:     ociCommon.String(signingDetails.KmsKeyVersionId),
			SigningAlgorithm: algorithm,
			MessageType:      keymanagement.SignDataDetailsMessageTypeRaw,
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to sign image due to %s", err)
	}
	return encoded, *signResponse.Signature, nil
}

func createImageSignatureMessage(region string, imageDigest string, repositoryName string, signingDetails common.SigningDetails) (string, error) {
	message := Message{
		Description:      "image signed by fn CLI",
		ImageDigest:      imageDigest,
		KmsKeyId:         signingDetails.KmsKeyId,
		KmsKeyVersionId:  signingDetails.KmsKeyVersionId,
		Region:           region,
		RepositoryName:   repositoryName,
		SigningAlgorithm: signingDetails.SigningAlgorithm,
		Metadata:         "{\"signedBy\":\"fn CLI\"}",
	}
	messageBytes, err := json.Marshal(&message)
	encoded := base64.StdEncoding.EncodeToString(messageBytes)
	if err != nil {
		return "", fmt.Errorf("failed to serialize image signature message due to %s", err)
	}
	return encoded, nil
}

func buildCryptoEndpoint(region, kmsKeyId string) (string, error) {
	keyIdRegexp := regexp.MustCompile(`ocid1\.key\.([\w-]+)\.([\w-]+)\.([\w-]+)\.([\w]{60})`)
	matches := keyIdRegexp.FindStringSubmatch(kmsKeyId)
	if len(matches) != 5 {
		return "", fmt.Errorf("keyId %s cannot be parsed", kmsKeyId)
	}
	vaultExt := matches[3]
	cryptoEndpointTemplate := "https://{vaultExt}-crypto.kms.{region}.oci.{secondLevelDomain}"
	ociRegion := ociCommon.StringToRegion(region)
	if _, ok := RegionsWithOldKMSEndpoints[ociRegion]; ok {
		cryptoEndpointTemplate = strings.Replace(cryptoEndpointTemplate, "oci.{secondLevelDomain}", "{secondLevelDomain}", -1)
	}
	cryptoEndpoint := ociRegion.EndpointForTemplate("kms", cryptoEndpointTemplate)
	return strings.Replace(cryptoEndpoint, "{vaultExt}", vaultExt, 1), nil
}

func uploadImageSignature(artifactsClient artifacts.ArtifactsClient, compartmentId string, imageId string, message string, signature string, signingDetails common.SigningDetails) error {
	algorithm := artifacts.CreateContainerImageSignatureDetailsSigningAlgorithmEnum(signingDetails.SigningAlgorithm)
	_, err := artifactsClient.CreateContainerImageSignature(context.Background(), artifacts.CreateContainerImageSignatureRequest{
		CreateContainerImageSignatureDetails: artifacts.CreateContainerImageSignatureDetails{
			CompartmentId:    ociCommon.String(compartmentId),
			ImageId:          ociCommon.String(imageId),
			KmsKeyId:         ociCommon.String(signingDetails.KmsKeyId),
			KmsKeyVersionId:  ociCommon.String(signingDetails.KmsKeyVersionId),
			SigningAlgorithm: algorithm,
			Message:          ociCommon.String(message),
			Signature:        ociCommon.String(signature),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to upload image signature due to %s", err)
	}
	return nil
}

func findMissingValues(signingDetails common.SigningDetails) string {
	var missingValues []string
	if signingDetails.ImageCompartmentId == "" {
		missingValues = append(missingValues, "image_compartment_id")
	}
	if signingDetails.KmsKeyId == "" {
		missingValues = append(missingValues, "kms_key_id")
	}
	if signingDetails.KmsKeyVersionId == "" {
		missingValues = append(missingValues, "kms_key_version_id")
	}
	if signingDetails.SigningAlgorithm == "" {
		missingValues = append(missingValues, "signing_algorithm")
	}
	return strings.Join(missingValues, ",")
}

// Checks if the repostitory exists in the compartment
func doesRepositoryExistInCompartment(repositoryName string, compartmentID string, artifactsClient artifacts.ArtifactsClient) (bool, error) {
	response, err := artifactsClient.ListContainerRepositories(context.Background(), artifacts.ListContainerRepositoriesRequest{
		CompartmentId: &compartmentID,
		DisplayName:   &repositoryName})
	if err != nil {
		return false, fmt.Errorf("failed to lookup container repository due to %w", err)
	}
	if *response.RepositoryCount == 1 {
		return true, nil
	}
	return false, nil
}

// This function tries to create the repository in compartmentID
func createContainerRepositoryInCompartment(repositoryName string, compartmentID string, artifactsClient artifacts.ArtifactsClient) error {
	_, err := artifactsClient.CreateContainerRepository(context.Background(), artifacts.CreateContainerRepositoryRequest{
		CreateContainerRepositoryDetails: artifacts.CreateContainerRepositoryDetails{
			CompartmentId: &compartmentID,
			DisplayName:   &repositoryName,
		},
	})
	return err
}
