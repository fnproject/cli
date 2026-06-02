package pbf

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	cliClient "github.com/fnproject/cli/client"
	fnprovideroracle "github.com/fnproject/fn_go/provider/oracle"
	ocifunctions "github.com/oracle/oci-go-sdk/v65/functions"
	"github.com/urfave/cli"
)

type pbfCmd struct {
	provider *fnprovideroracle.OracleProvider
	client   *ocifunctions.FunctionsManagementClient
}

func buildFunctionsManagementClient(provider *fnprovideroracle.OracleProvider) (*ocifunctions.FunctionsManagementClient, error) {
	client, err := ocifunctions.NewFunctionsManagementClientWithConfigurationProvider(provider.ConfigurationProvider)
	if err != nil {
		return nil, err
	}
	if provider.FnApiUrl != nil {
		client.Host = provider.FnApiUrl.String()
	} else {
		region, _ := provider.ConfigurationProvider.Region()
		if region != "" {
			client.SetRegion(region)
		}
	}
	return &client, nil
}

func initPBFClient() (*pbfCmd, error) {
	provider, err := cliClient.CurrentProvider()
	if err != nil {
		return nil, err
	}
	oracleProvider, ok := provider.(*fnprovideroracle.OracleProvider)
	if !ok || oracleProvider == nil {
		return nil, fmt.Errorf("PBF commands are only supported with an oracle provider")
	}
	mgmtClient, err := buildFunctionsManagementClient(oracleProvider)
	if err != nil {
		return nil, err
	}
	return &pbfCmd{provider: oracleProvider, client: mgmtClient}, nil
}

func formatSDKTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func formatListingTriggers(triggers []ocifunctions.Trigger) string {
	if len(triggers) == 0 {
		return ""
	}
	names := make([]string, 0, len(triggers))
	for _, trig := range triggers {
		if trig.Name != nil {
			names = append(names, *trig.Name)
		}
	}
	return strings.Join(names, ",")
}

func isOCID(value string) bool {
	return strings.HasPrefix(strings.TrimSpace(value), "ocid1.")
}

func (p *pbfCmd) resolveListingID(identifier string) (string, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return "", fmt.Errorf("missing PBF listing identifier")
	}
	if isOCID(identifier) {
		return identifier, nil
	}
	limit := 1
	req := ocifunctions.ListPbfListingsRequest{Name: &identifier, Limit: &limit}
	res, err := p.client.ListPbfListings(context.Background(), req)
	if err != nil {
		return "", err
	}
	if len(res.Items) == 0 || res.Items[0].Id == nil {
		return "", fmt.Errorf("PBF listing %q not found", identifier)
	}
	return *res.Items[0].Id, nil
}

func printListings(c *cli.Context, items []ocifunctions.PbfListingSummary) error {
	if strings.EqualFold(c.String("output"), "json") {
		b, err := json.MarshalIndent(items, "", "    ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(os.Stdout, string(b))
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "NAME\tPUBLISHER\tUPDATED\tTRIGGERS\tID")
	for _, item := range items {
		name := ""
		if item.Name != nil {
			name = *item.Name
		}
		publisher := ""
		if item.PublisherDetails != nil && item.PublisherDetails.Name != nil {
			publisher = *item.PublisherDetails.Name
		}
		updated := ""
		if item.TimeUpdated != nil {
			t := item.TimeUpdated.Time
			updated = formatSDKTime(&t)
		}
		id := ""
		if item.Id != nil {
			id = *item.Id
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, publisher, updated, formatListingTriggers(item.Triggers), id)
	}
	return w.Flush()
}

func printListingVersions(c *cli.Context, items []ocifunctions.PbfListingVersionSummary) error {
	if strings.EqualFold(c.String("output"), "json") {
		b, err := json.MarshalIndent(items, "", "    ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(os.Stdout, string(b))
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "VERSION\tSTATE\tUPDATED\tMIN_MEMORY_MBS\tID")
	for _, item := range items {
		version := ""
		if item.Name != nil {
			version = *item.Name
		}
		updated := ""
		if item.TimeUpdated != nil {
			t := item.TimeUpdated.Time
			updated = formatSDKTime(&t)
		}
		minMemory := ""
		if item.Requirements != nil && item.Requirements.MinMemoryRequiredInMBs != nil {
			minMemory = fmt.Sprintf("%d", *item.Requirements.MinMemoryRequiredInMBs)
		}
		id := ""
		if item.Id != nil {
			id = *item.Id
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", version, item.LifecycleState, updated, minMemory, id)
	}
	return w.Flush()
}

func printPBFTriggers(c *cli.Context, items []ocifunctions.TriggerSummary) error {
	if strings.EqualFold(c.String("output"), "json") {
		b, err := json.MarshalIndent(items, "", "    ")
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(os.Stdout, string(b))
		return err
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	_, _ = fmt.Fprintln(w, "NAME")
	for _, item := range items {
		name := ""
		if item.Name != nil {
			name = *item.Name
		}
		_, _ = fmt.Fprintf(w, "%s\n", name)
	}
	return w.Flush()
}

func (p *pbfCmd) listListings(c *cli.Context) error {
	limit := c.Int("limit")
	req := ocifunctions.ListPbfListingsRequest{}
	if limit > 0 {
		req.Limit = &limit
	}
	if search := strings.TrimSpace(c.String("search")); search != "" {
		req.NameContains = &search
	}
	if trigger := strings.TrimSpace(c.String("trigger")); trigger != "" {
		req.Trigger = []string{trigger}
	}
	res, err := p.client.ListPbfListings(context.Background(), req)
	if err != nil {
		return err
	}
	return printListings(c, res.Items)
}

func (p *pbfCmd) getListing(c *cli.Context) error {
	listingID, err := p.resolveListingID(c.Args().First())
	if err != nil {
		return err
	}
	res, err := p.client.GetPbfListing(context.Background(), ocifunctions.GetPbfListingRequest{PbfListingId: &listingID})
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(res.PbfListing, "", "    ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(os.Stdout, string(b))
	return err
}

func (p *pbfCmd) listVersions(c *cli.Context) error {
	listingID, err := p.resolveListingID(c.Args().First())
	if err != nil {
		return err
	}
	limit := c.Int("limit")
	req := ocifunctions.ListPbfListingVersionsRequest{PbfListingId: &listingID}
	if limit > 0 {
		req.Limit = &limit
	}
	if c.Bool("current") {
		current := true
		req.IsCurrentVersion = &current
	}
	res, err := p.client.ListPbfListingVersions(context.Background(), req)
	if err != nil {
		return err
	}
	return printListingVersions(c, res.Items)
}

func (p *pbfCmd) getVersion(c *cli.Context) error {
	versionID := strings.TrimSpace(c.Args().First())
	if versionID == "" {
		return fmt.Errorf("missing PBF listing version identifier")
	}
	res, err := p.client.GetPbfListingVersion(context.Background(), ocifunctions.GetPbfListingVersionRequest{PbfListingVersionId: &versionID})
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(res.PbfListingVersion, "", "    ")
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(os.Stdout, string(b))
	return err
}

func (p *pbfCmd) listTriggers(c *cli.Context) error {
	limit := c.Int("limit")
	req := ocifunctions.ListTriggersRequest{}
	if limit > 0 {
		req.Limit = &limit
	}
	if name := strings.TrimSpace(c.Args().First()); name != "" {
		req.Name = &name
	}
	res, err := p.client.ListTriggers(context.Background(), req)
	if err != nil {
		return err
	}
	return printPBFTriggers(c, res.Items)
}
