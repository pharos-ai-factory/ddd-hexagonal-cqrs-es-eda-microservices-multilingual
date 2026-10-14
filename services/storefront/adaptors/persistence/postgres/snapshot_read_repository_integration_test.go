//go:build integration

package postgres

import (
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	"reflect"
	"testing"
)

func TestQueriesTraverseMoreThanOnePageAndRetainUnpaginatedReads(t *testing.T) {
	db, owner := fixtureDB(t), admin(t)
	commands, queries := snapshotDecisions[counter](db, "test_query_counter"), NewSnapshotReadRepository[counter](db, "test_query_counter")
	for index := 0; index < 105; index++ {
		id := NewID()
		clean(t, owner, id)
		_, err := commands.Execute(t.Context(), metadata(id, 0), func(a.Loaded[counter]) (a.Mutation[counter], error) { return a.Changed(counter{index}, "active"), nil })
		if err != nil {
			t.Fatal(err)
		}
	}
	all, err := queries.List(t.Context())
	if err != nil || len(all) < 105 {
		t.Fatal(len(all), err)
	}
	var combined []a.Loaded[counter]
	request := a.PageRequest{Limit: 37}
	for {
		page, err := queries.Page(t.Context(), request)
		if err != nil {
			t.Fatal(err)
		}
		if len(page.Items) > request.Limit {
			t.Fatal("unbounded page")
		}
		combined = append(combined, page.Items...)
		if page.NextID == "" {
			break
		}
		if page.NextID <= request.After {
			t.Fatal("cursor did not advance")
		}
		request.After = page.NextID
	}
	if !reflect.DeepEqual(all, combined) {
		t.Fatal("pagination lost, repeated or reordered rows")
	}
}
