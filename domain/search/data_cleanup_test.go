package search

import (
	"testing"
)

func TestCleanupStringWithGithubIssueTemplate(t *testing.T) {
	actual := CleanupString(
		// https://github.com/shopware/platform/issues/2445
		"### PHP Version\n\n7\n\n### Shopware Version\n\n6.4.10\n\n### Expected behaviour\n\nPaging in Custom-Field-Sets works\n\n### Actual behaviour\n\npaging doesnt work. Always Page 1 is requested.\n\n### How to reproduce\n\ncreate more than 25 custom-field-sets and try to page to page 2.",
	)

	expected := "Paging in Custom-Field-Sets works\n\npaging doesnt work. Always Page 1 is requested.\n\ncreate more than 25 custom-field-sets and try to page to page 2."

	if actual != expected {
		t.Errorf("Expected cleaned up string to be '%s' but got '%s'", expected, actual)
	}
}

func TestCleanupStringWithIssueTrackerTemplate(t *testing.T) {
	actual := CleanupString(
		// https://issues.shopware.com/issues/NEXT-25647
		"Description: Inside the entity.xml file (app system) entities can be described and on install the corresponding database tables are created. However, if the app is uninstalled and \"Remove all app data permanently\" is *not* selected the database tables are removed anyway.\n\nEnvironment: 6.4.20.0\n\nSteps to reproduce: Create an entity.xml file within a valid app. Install the app. Then uninstall the app, but *don't* select \"Remove all app data permanently\".\n\nExpected result: The database tables still exist\n\nActual result: The database tables are removed\n\n--------------------\n\nThis way essential data is lost even though the user selected to keep the user data.",
	)

	expected := "Inside the entity.xml file (app system) entities can be described and on install the corresponding database tables are created. However, if the app is uninstalled and \"Remove all app data permanently\" is not selected the database tables are removed anyway.\n\n6.4.20.0\n\nCreate an entity.xml file within a valid app. Install the app. Then uninstall the app, but don't select \"Remove all app data permanently\".\n\nThe database tables still exist\n\nThe database tables are removed\n\n--------------------\n\nThis way essential data is lost even though the user selected to keep the user data."

	if actual != expected {
		t.Errorf("Expected cleaned up string to be '%s' but got '%s'", expected, actual)
	}
}

func TestCleanupStringWithGithubPRTemplate(t *testing.T) {
	actual := CleanupString(
		// https://issues.shopware.com/issues/NEXT-9855
		"<!--\nThank you for contributing to Shopware! Please fill out this description template to help us to process your pull request.\n\nPlease make sure to fulfil our contribution guideline (https://docs.shopware.com/en/shopware-platform-dev-en/community/contribution-guideline?category=shopware-platform-dev-en/community).\n\nDo your changes need to be mentioned in the documentation?\nAdd notes on your change right now in the documentation files in /src/Docs/Resources and add them to the pull request as well. \n-->\n\n### 1. Why is this change necessary?\nNo more than 10 custom fields can be handled in the Custom Field Set administration view.\n\n### 2. What does this change do, exactly?\nAdd the total-count-mode param.\n\n### 3. Describe each step to reproduce the issue or behaviour.\nIf you are on the detail page of a customfieldset you only get 10 customfields displayed, because there's a parameter limit: 10 when calling the API.\n\nAs there's no pagination, that makes it impossible to get the other custom fields.\n\n### 4. Please link to the relevant issues (if any).\n#1105 \n\n### 5. Checklist\n\n- [ ] I have written tests and verified that they fail without my change\n- [x] I have squashed any insignificant commits\n- [x] I have written or adjusted the documentation according to my changes\n- [x] This change has comments for package types, values, functions, and non-obvious lines of code\n- [x] I have read the contribution requirements and fulfil them.\n",
	)

	expected := "No more than 10 custom fields can be handled in the Custom Field Set administration view.\n\nAdd the total-count-mode param.\n\nIf you are on the detail page of a customfieldset you only get 10 customfields displayed, because there's a parameter limit: 10 when calling the API.\n\nAs there's no pagination, that makes it impossible to get the other custom fields.\n\n \n1105"

	if actual != expected {
		t.Errorf("Expected cleaned up string to be '%s' but got '%s'", expected, actual)
	}
}

func TestCleanupStringWithOnlyNewlinesTemplate(t *testing.T) {
	actual := CleanupString(
		"\n\n",
	)

	if actual != "" {
		t.Errorf("Expected cleaned up string to be '\"\"' but got '%s'", actual)
	}
}

func TestCleanupStringWithEmptyGithubTemplate(t *testing.T) {
	actual := CleanupString(
		// https://github.com/shopware/platform/pull/1281
		"<!--\nThank you for contributing to Shopware! Please fill out this description template to help us to process your pull request.\n\nPlease make sure to fulfil our contribution guideline (https://docs.shopware.com/en/shopware-platform-dev-en/contribution/contribution-guideline?category=shopware-platform-dev-en/contribution).\n\nDo your changes need to be mentioned in the documentation?\nAdd notes on your change right now in the documentation files in /src/Docs/Resources and add them to the pull request as well. \n-->\n\n### 1. Why is this change necessary?\n\n\n### 2. What does this change do, exactly?\n\n\n### 3. Describe each step to reproduce the issue or behaviour.\n\n\n### 4. Please link to the relevant issues (if any).\n\n\n### 5. Checklist\n\n- [ ] I have written tests and verified that they fail without my change\n- [ ] I have squashed any insignificant commits\n- [ ] I have written or adjusted the documentation according to my changes\n- [ ] This change has comments for package types, values, functions, and non-obvious lines of code\n- [ ] I have read the contribution requirements and fulfil them.\n",
	)

	if actual != "" {
		t.Errorf("Expected cleaned up string to be '\"\"' but got '%s'", actual)
	}
}
