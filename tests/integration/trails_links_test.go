// CLI integration tests for trails, links, and archive operations.
// These tests do NOT belong in test001-self-hosting; they will be extracted
// into a separate test suite by Task 2.
// Implements: crumbs-ag8.1 (convert validation script to Go tests).
package integration

import (
	"path/filepath"
	"testing"
)

// TestCreateTrails verifies trail creation.
func TestCreateTrails(t *testing.T) {
	env := NewTestEnv(t)
	env.MustRunCupboard("init")

	// Create first trail in active state
	result1 := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail1 := ParseJSON[Trail](t, result1.Stdout)
	if trail1.TrailID == "" {
		t.Error("trail1 ID not generated")
	}
	if trail1.State != "active" {
		t.Errorf("trail1 state mismatch: got %q", trail1.State)
	}

	// Create second trail
	result2 := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail2 := ParseJSON[Trail](t, result2.Stdout)
	if trail2.TrailID == "" {
		t.Error("trail2 ID not generated")
	}

	// Verify different IDs
	if trail1.TrailID == trail2.TrailID {
		t.Error("trail IDs should be unique")
	}
}

// TestLinkCrumbsToTrails verifies belongs_to link creation.
func TestLinkCrumbsToTrails(t *testing.T) {
	env := NewTestEnv(t)
	env.MustRunCupboard("init")

	// Create crumbs
	crumb1Result := env.MustRunCupboard("set", "crumbs", "", `{"Name":"Crumb for trail 1","State":"draft"}`)
	crumb1 := ParseJSON[Crumb](t, crumb1Result.Stdout)

	crumb2Result := env.MustRunCupboard("set", "crumbs", "", `{"Name":"Crumb for trail 2","State":"draft"}`)
	crumb2 := ParseJSON[Crumb](t, crumb2Result.Stdout)

	// Create trails
	trail1Result := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail1 := ParseJSON[Trail](t, trail1Result.Stdout)

	trail2Result := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail2 := ParseJSON[Trail](t, trail2Result.Stdout)

	// Link crumb1 to trail1
	link1JSON := `{"LinkType":"belongs_to","FromID":"` + crumb1.CrumbID + `","ToID":"` + trail1.TrailID + `"}`
	link1Result := env.MustRunCupboard("set", "links", "", link1JSON)
	link1 := ParseJSON[Link](t, link1Result.Stdout)
	if link1.LinkID == "" {
		t.Error("link1 ID not generated")
	}

	// Link crumb2 to trail2
	link2JSON := `{"LinkType":"belongs_to","FromID":"` + crumb2.CrumbID + `","ToID":"` + trail2.TrailID + `"}`
	link2Result := env.MustRunCupboard("set", "links", "", link2JSON)
	link2 := ParseJSON[Link](t, link2Result.Stdout)
	if link2.LinkID == "" {
		t.Error("link2 ID not generated")
	}

	// Verify links exist
	linksResult := env.MustRunCupboard("list", "links")
	links := ParseJSON[[]Link](t, linksResult.Stdout)
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}
}

// TestCompleteTrail verifies trail completion (successful exploration).
func TestCompleteTrail(t *testing.T) {
	env := NewTestEnv(t)
	env.MustRunCupboard("init")

	// Create trail in active state
	trailResult := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail := ParseJSON[Trail](t, trailResult.Stdout)

	// Complete trail
	env.MustRunCupboard("set", "trails", trail.TrailID,
		`{"TrailID":"`+trail.TrailID+`","State":"completed"}`)

	// Verify state
	getResult := env.MustRunCupboard("get", "trails", trail.TrailID)
	trail = ParseJSON[Trail](t, getResult.Stdout)
	if trail.State != "completed" {
		t.Errorf("expected trail state completed, got %q", trail.State)
	}
}

// TestAbandonTrail verifies trail abandonment (failed exploration).
func TestAbandonTrail(t *testing.T) {
	env := NewTestEnv(t)
	env.MustRunCupboard("init")

	// Create trail in active state
	trailResult := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail := ParseJSON[Trail](t, trailResult.Stdout)

	// Abandon trail
	env.MustRunCupboard("set", "trails", trail.TrailID,
		`{"TrailID":"`+trail.TrailID+`","State":"abandoned"}`)

	// Verify state
	getResult := env.MustRunCupboard("get", "trails", trail.TrailID)
	trail = ParseJSON[Trail](t, getResult.Stdout)
	if trail.State != "abandoned" {
		t.Errorf("expected trail state abandoned, got %q", trail.State)
	}
}

// TestArchiveCrumb verifies crumb archival (soft delete via dust state).
func TestArchiveCrumb(t *testing.T) {
	env := NewTestEnv(t)
	env.MustRunCupboard("init")

	// Create crumbs
	crumb1Result := env.MustRunCupboard("set", "crumbs", "", `{"Name":"Crumb to keep","State":"draft"}`)
	ParseJSON[Crumb](t, crumb1Result.Stdout)

	crumb2Result := env.MustRunCupboard("set", "crumbs", "", `{"Name":"Crumb to dust","State":"draft"}`)
	crumb2 := ParseJSON[Crumb](t, crumb2Result.Stdout)

	// Dust crumb2 (mark as failed/abandoned)
	env.MustRunCupboard("set", "crumbs", crumb2.CrumbID,
		`{"CrumbID":"`+crumb2.CrumbID+`","Name":"Crumb to dust","State":"dust"}`)

	// Verify state
	getResult := env.MustRunCupboard("get", "crumbs", crumb2.CrumbID)
	crumb2 = ParseJSON[Crumb](t, getResult.Stdout)
	if crumb2.State != "dust" {
		t.Errorf("expected crumb state dust, got %q", crumb2.State)
	}

	// Verify dust crumb not in draft list
	draftResult := env.MustRunCupboard("list", "crumbs", "State=draft")
	draftCrumbs := ParseJSON[[]Crumb](t, draftResult.Stdout)
	if len(draftCrumbs) != 1 {
		t.Errorf("expected 1 draft crumb after dust, got %d", len(draftCrumbs))
	}
}

// TestTrailsPersistence verifies trails are persisted to JSONL files.
func TestTrailsPersistence(t *testing.T) {
	env := NewTestEnv(t)
	env.MustRunCupboard("init")

	// Create test data
	env.MustRunCupboard("set", "crumbs", "", `{"Name":"Crumb 1","State":"draft"}`)
	env.MustRunCupboard("set", "crumbs", "", `{"Name":"Crumb 2","State":"draft"}`)
	env.MustRunCupboard("set", "crumbs", "", `{"Name":"Crumb 3","State":"draft"}`)

	trail1Result := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail1 := ParseJSON[Trail](t, trail1Result.Stdout)

	trail2Result := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail2 := ParseJSON[Trail](t, trail2Result.Stdout)

	// Complete trail1, abandon trail2
	env.MustRunCupboard("set", "trails", trail1.TrailID,
		`{"TrailID":"`+trail1.TrailID+`","State":"completed"}`)
	env.MustRunCupboard("set", "trails", trail2.TrailID,
		`{"TrailID":"`+trail2.TrailID+`","State":"abandoned"}`)

	// Verify trails.jsonl
	trailsFile := filepath.Join(env.DataDir, "trails.jsonl")
	trails := ReadJSONLFile[map[string]any](t, trailsFile)

	// Verify trail states in JSON
	for _, tr := range trails {
		trailID, _ := tr["trail_id"].(string)
		state, _ := tr["state"].(string)
		if trailID == trail1.TrailID && state != "completed" {
			t.Errorf("expected trail1 state completed in JSON, got %q", state)
		}
		if trailID == trail2.TrailID && state != "abandoned" {
			t.Errorf("expected trail2 state abandoned in JSON, got %q", state)
		}
	}
}

// TestFullWorkflowWithTrails verifies the complete workflow with trails.
func TestFullWorkflowWithTrails(t *testing.T) {
	env := NewTestEnv(t)
	env.MustRunCupboard("init")

	// Create 3 crumbs
	crumb1Result := env.MustRunCupboard("set", "crumbs", "", `{"Name":"Implement feature X","State":"draft"}`)
	crumb1 := ParseJSON[Crumb](t, crumb1Result.Stdout)

	crumb2Result := env.MustRunCupboard("set", "crumbs", "", `{"Name":"Write tests for feature X","State":"draft"}`)
	crumb2 := ParseJSON[Crumb](t, crumb2Result.Stdout)

	crumb3Result := env.MustRunCupboard("set", "crumbs", "", `{"Name":"Try approach A","State":"draft"}`)
	crumb3 := ParseJSON[Crumb](t, crumb3Result.Stdout)

	// Transition crumb1 through states to pebble (completed successfully)
	env.MustRunCupboard("set", "crumbs", crumb1.CrumbID,
		`{"CrumbID":"`+crumb1.CrumbID+`","Name":"Implement feature X","State":"pebble"}`)

	// Create 2 trails
	trail1Result := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail1 := ParseJSON[Trail](t, trail1Result.Stdout)

	trail2Result := env.MustRunCupboard("set", "trails", "", `{"State":"active"}`)
	trail2 := ParseJSON[Trail](t, trail2Result.Stdout)

	// Link crumbs to trails
	env.MustRunCupboard("set", "links", "",
		`{"LinkType":"belongs_to","FromID":"`+crumb2.CrumbID+`","ToID":"`+trail1.TrailID+`"}`)
	env.MustRunCupboard("set", "links", "",
		`{"LinkType":"belongs_to","FromID":"`+crumb3.CrumbID+`","ToID":"`+trail2.TrailID+`"}`)

	// Complete trail1, abandon trail2
	env.MustRunCupboard("set", "trails", trail1.TrailID,
		`{"TrailID":"`+trail1.TrailID+`","State":"completed"}`)
	env.MustRunCupboard("set", "trails", trail2.TrailID,
		`{"TrailID":"`+trail2.TrailID+`","State":"abandoned"}`)

	// Dust crumb3 (from abandoned trail - mark as failed/abandoned)
	env.MustRunCupboard("set", "crumbs", crumb3.CrumbID,
		`{"CrumbID":"`+crumb3.CrumbID+`","Name":"Try approach A","State":"dust"}`)

	// Final validation - counts
	allCrumbs := ParseJSON[[]Crumb](t, env.MustRunCupboard("list", "crumbs").Stdout)
	allTrails := ParseJSON[[]Trail](t, env.MustRunCupboard("list", "trails").Stdout)
	allLinks := ParseJSON[[]Link](t, env.MustRunCupboard("list", "links").Stdout)

	if len(allCrumbs) != 3 {
		t.Errorf("expected 3 crumbs, got %d", len(allCrumbs))
	}
	if len(allTrails) != 2 {
		t.Errorf("expected 2 trails, got %d", len(allTrails))
	}
	if len(allLinks) != 2 {
		t.Errorf("expected 2 links, got %d", len(allLinks))
	}

	// State counts
	pebbleCrumbs := ParseJSON[[]Crumb](t, env.MustRunCupboard("list", "crumbs", "State=pebble").Stdout)
	dustCrumbs := ParseJSON[[]Crumb](t, env.MustRunCupboard("list", "crumbs", "State=dust").Stdout)
	draftCrumbs := ParseJSON[[]Crumb](t, env.MustRunCupboard("list", "crumbs", "State=draft").Stdout)
	completedTrails := ParseJSON[[]Trail](t, env.MustRunCupboard("list", "trails", "State=completed").Stdout)
	abandonedTrails := ParseJSON[[]Trail](t, env.MustRunCupboard("list", "trails", "State=abandoned").Stdout)

	if len(pebbleCrumbs) != 1 {
		t.Errorf("expected 1 pebble crumb, got %d", len(pebbleCrumbs))
	}
	if len(dustCrumbs) != 1 {
		t.Errorf("expected 1 dust crumb, got %d", len(dustCrumbs))
	}
	if len(draftCrumbs) != 1 {
		t.Errorf("expected 1 draft crumb, got %d", len(draftCrumbs))
	}
	if len(completedTrails) != 1 {
		t.Errorf("expected 1 completed trail, got %d", len(completedTrails))
	}
	if len(abandonedTrails) != 1 {
		t.Errorf("expected 1 abandoned trail, got %d", len(abandonedTrails))
	}
}
