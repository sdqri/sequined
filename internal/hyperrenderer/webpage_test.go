package hyperrenderer_test

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"testing"

	"github.com/goccy/go-graphviz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	hr "github.com/sdqri/sequined/internal/hyperrenderer"
)

func TestClone(t *testing.T) {
	webpage := hr.NewWebpage(hr.WebpageTypeAuthority)

	cloned := webpage.Clone(hr.WebpageTypeHub)

	assert.Equal(t, webpage.AuthorityTmpl, cloned.AuthorityTmpl, "AuthorityTmpl changed after cloning")
	assert.Equal(t, webpage.HubTmpl, cloned.HubTmpl, "HubTmpl changed after cloning")
}

func TestAddChild(t *testing.T) {
	root := hr.NewWebpage(hr.WebpageTypeAuthority)

	child := root.AddChild(hr.WebpageTypeHub)

	assert.Len(t, root.Links, 1, "Parent page should have 1 link after adding child")
	assert.Equal(t, child, root.Links[0], "Child page should be added to parent's links")

	assert.Equal(t, root, child.Parent, "Child page's parent pointer should point to parent page")
}

//	func TestGetID(t *testing.T) {
//		testCases := []struct {
//			description string
//			webpage     *Webpage
//			expectedID  string
//		}{
//			{
//				description: "Webpage with ID 1",
//				webpage:     &Webpage{ID: 1},
//				expectedID:  "1",
//			},
//			// Add more test cases as needed
//		}
//
//		for _, tc := range testCases {
//			t.Run(tc.description, func(t *testing.T) {
//				assert.Equal(t, tc.expectedID, tc.webpage.GetID(), "Unexpected ID returned")
//			})
//		}
//	}

func TestHardCodedGetPath(t *testing.T) {
	parentPage := &hr.Webpage{ID: 1, Path: "/", Type: hr.WebpageTypeAuthority}
	childPage := &hr.Webpage{ID: 2, Path: "/2", Parent: parentPage, Type: hr.WebpageTypeHub}
	grandChildPage := &hr.Webpage{ID: 3, Path: "/2/3", Parent: childPage, Type: hr.WebpageTypeHub}

	testCases := []struct {
		description  string
		webpage      *hr.Webpage
		expectedPath string
	}{
		{
			description:  "Root Webpage",
			webpage:      parentPage,
			expectedPath: "/",
		},
		{
			description:  "Child webpage",
			webpage:      childPage,
			expectedPath: "/2",
		},
		{
			description:  "Grandchild webpage",
			webpage:      grandChildPage,
			expectedPath: "/2/3",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedPath, tc.webpage.GetPath(), "Unexpected path returned")
		})
	}
}

func TestGetPathWithDefaultPathGenerator(t *testing.T) {
	parentPage := hr.NewWebpage(hr.WebpageTypeHub)
	childPage := parentPage.AddChild(hr.WebpageTypeHub)
	grandChildPage := childPage.AddChild(hr.WebpageTypeHub)

	testCases := []struct {
		description  string
		webpage      *hr.Webpage
		expectedPath string
	}{
		{
			description:  "Root Webpage",
			webpage:      parentPage,
			expectedPath: "/",
		},
		{
			description:  "Child webpage",
			webpage:      childPage,
			expectedPath: fmt.Sprintf("/%s", childPage.GetID()),
		},
		{
			description:  "Grandchild webpage",
			webpage:      grandChildPage,
			expectedPath: fmt.Sprintf("/%s/%s", childPage.GetID(), grandChildPage.GetID()),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expectedPath, tc.webpage.GetPath(), "Unexpected path returned")
		})
	}
}

func TestRender(t *testing.T) {
	authorityTemplate := `Authority Template: {{ .Node.ID }}`
	hubTemplate := `Hub Template: {{ .Node.ID }}`
	customTemplate := `Custom Template: {{ .Node.ID }}`

	authorityTmpl, err := template.New("authority").Parse(authorityTemplate)
	if err != nil {
		t.Fatalf("Error parsing authority template: %v", err)
	}
	hubTmpl, err := template.New("hub").Parse(hubTemplate)
	if err != nil {
		t.Fatalf("Error parsing hub template: %v", err)
	}
	customTmpl, err := template.New("custom").Parse(customTemplate)
	if err != nil {
		t.Fatalf("Error parsing custom template: %v", err)
	}

	root := hr.NewWebpage(hr.WebpageTypeHub, hr.WithHubTemplate(hubTmpl), hr.WithAuthorityTemplate(authorityTmpl))
	child := root.AddChild(hr.WebpageTypeAuthority)
	grandChild := child.AddChild(hr.WebpageTypeAuthority, hr.WithCustomTemplate(customTmpl))

	testCases := []struct {
		name           string
		webpage        *hr.Webpage
		expectedOutput string
	}{
		{
			name:           "Render root Page",
			webpage:        root,
			expectedOutput: "Hub Template: %s",
		},
		{
			name:           "Render child Page",
			webpage:        child,
			expectedOutput: "Authority Template: %s",
		},
		{
			name:           "Render grandChild Page",
			webpage:        grandChild,
			expectedOutput: "Custom Template: %s",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			err := tc.webpage.Render(&buf)
			if err != nil {
				t.Fatalf("Error rendering webpage: %v", err)
			}

			// Verify the output matches the expected output
			assert.Equal(t, fmt.Sprintf(tc.expectedOutput, tc.webpage.GetID()), buf.String(), "Unexpected output")
		})
	}
}

func TestGetLinks(t *testing.T) {
	rootPage := &hr.Webpage{ID: 1}
	childPage2 := &hr.Webpage{ID: 2}
	childPage3 := &hr.Webpage{ID: 3}
	childPage4 := &hr.Webpage{ID: 4}

	rootPage.AddLink(childPage2)
	childPage2.AddLink(childPage3)
	childPage2.AddLink(childPage4)

	testCases := []struct {
		description       string
		webpage           *hr.Webpage
		expectedLinkCount int
	}{
		{
			description:       "No links",
			webpage:           childPage4,
			expectedLinkCount: 0,
		},
		{
			description:       "One link",
			webpage:           rootPage,
			expectedLinkCount: 1,
		},
		{
			description:       "two links",
			webpage:           childPage2,
			expectedLinkCount: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			links := tc.webpage.GetLinks()

			assert.Equal(t, tc.expectedLinkCount, len(links), "Unexpected number of links")
		})
	}
}

func TestAddLink(t *testing.T) {
	testCases := []struct {
		description         string
		parentWebpageType   hr.WebpageType
		childWebpageType    hr.WebpageType
		expectParentToChild bool
		expectChildInParent bool
	}{
		{
			description:         "Authority to Hub",
			parentWebpageType:   hr.WebpageTypeAuthority,
			childWebpageType:    hr.WebpageTypeHub,
			expectParentToChild: true,
			expectChildInParent: true,
		},
		{
			description:         "Hub to Authority",
			parentWebpageType:   hr.WebpageTypeHub,
			childWebpageType:    hr.WebpageTypeAuthority,
			expectParentToChild: true,
			expectChildInParent: true,
		},
		{
			description:         "Authority to Authority",
			parentWebpageType:   hr.WebpageTypeAuthority,
			childWebpageType:    hr.WebpageTypeAuthority,
			expectParentToChild: true,
			expectChildInParent: true,
		},
		{
			description:         "Hub to Hub",
			parentWebpageType:   hr.WebpageTypeHub,
			childWebpageType:    hr.WebpageTypeHub,
			expectParentToChild: true,
			expectChildInParent: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			parentPage := hr.NewWebpage(tc.parentWebpageType)
			childPage := hr.NewWebpage(tc.childWebpageType)

			parentPage.AddLink(childPage)

			if tc.expectParentToChild {
				assert.Equal(t, parentPage, childPage.Parent, "Incorrect parent assigned to child")
			} else {
				assert.Nil(t, childPage.Parent, "Child should not have a parent")
			}

			if tc.expectChildInParent {
				assert.Contains(t, parentPage.Links, childPage, "Child not added to parent's links")
			} else {
				assert.NotContains(t, parentPage.Links, childPage, "Child should not be in parent's links")
			}
		})
	}
}

var update = flag.Bool("update", false, "update golden files")

func TestDrawGolden(t *testing.T) {
	webpage := hr.Webpage{ID: 1}

	filename := "./test_output.dot"

	err := webpage.Draw(filename, graphviz.XDOT)
	require.NoError(t, err, "Error generating DOT file")

	actual, err := os.ReadFile(filename)
	require.NoError(t, err, "Error reading golden DOT file")

	goldenPath := filepath.Join("test-fixtures", "draw.golden")
	if *update {
		os.WriteFile(goldenPath, actual, 0644)
	}

	expected, err := os.ReadFile(goldenPath)
	require.NoError(t, err, "Error reading draw.golden file")

	assert.Equal(t, string(expected), string(actual), "Generated output does not match golden output")
	err = os.Remove(filename)
	require.NoError(t, err, "Error deleting generated DOT file")
}
