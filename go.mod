module github.com/mesh-intelligence/crumbs

go 1.25.7

replace github.com/mesh-intelligence/crumbs => ./

require (
	github.com/magefile/mage v1.15.0
	github.com/mesh-intelligence/mage-claude-orchestrator v0.0.0-20260214034930-18a029293594
)

require gopkg.in/yaml.v3 v3.0.1 // indirect
