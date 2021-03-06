package main

import (
	"errors"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
)

func isTagConstraintSpecificTag(tagConstraint string) (bool, string) {
	if len(tagConstraint) > 0 {
		switch tagConstraint[0] {
		// Check for a tagConstraint '='
		case '=':
			return true, strings.TrimSpace(tagConstraint[1:])

		// Check for a tagConstraint without constraint specifier
		// Neither of '!=', '>', '>=', '<', '<=', '~>' is prefixed before tag
		case '>', '<', '!', '~':
			return false, tagConstraint

		default:
			return true, strings.TrimSpace(tagConstraint)
		}
	}
	return false, tagConstraint
}

func getLatestAcceptableTag(tagConstraint string, tags []string) (string, *FetchError) {
	if len(tags) == 0 {
		return "", nil
	}

	// Sort all tags
	// Our use of the library go-version means that each tag will each be represented as a *version.Version
	// go-version normalizes the versions so store off a mapping from the normalized version back to the original tag.
	versions := make([]*version.Version, len(tags))
	verToTag := make(map[*version.Version]string)
	for i, tag := range tags {
		v, err := version.NewVersion(tag)
		if err != nil {
			return "", wrapError(err)
		}

		versions[i] = v
		verToTag[v] = tag
	}
	sort.Sort(version.Collection(versions))

	// If the tag constraint is empty, set it to the latest tag
	if tagConstraint == "" {
		tagConstraint = versions[len(versions)-1].String()
	}

	// Find the latest version that matches the given tag constraint
	constraints, err := version.NewConstraint(tagConstraint)
	if err != nil {
		// Explicitly check for a malformed tag value so we can return a nice error to the user
		if strings.Contains(err.Error(), "Malformed constraint") {
			return "", newError(invalidTagConstraintExpression, err.Error())
		} else {
			return "", wrapError(err)
		}
	}

	latestAcceptableVersion := versions[0]
	for _, version := range versions {
		if constraints.Check(version) && version.GreaterThan(latestAcceptableVersion) {
			latestAcceptableVersion = version
		}
	}

	// check constraint against latest acceptable version
	if !constraints.Check(latestAcceptableVersion) {
		return "", wrapError(errors.New("Tag does not exist"))
	}

	return verToTag[latestAcceptableVersion], nil
}
