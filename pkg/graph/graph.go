// pkg/graph/graph.go
package graph

import "buildy/pkg/config"

type Graph map[string][]string

func BuildDependencyGraph(subProjects []config.SubProject) Graph {
	graph := make(Graph)
	for _, subProject := range subProjects {
		graph[subProject.Name] = subProject.DependsOn
	}
	return graph
}

func (g Graph) GetSubprojectsToBuild(changedSubprojects []string) []string {
	var result []string
	visited := make(map[string]bool)

	var dfs func(subProject string)
	dfs = func(subProject string) {
		if visited[subProject] {
			return
		}
		visited[subProject] = true
		result = append(result, subProject)

		for _, dependency := range g[subProject] {
			dfs(dependency)
		}
	}

	for _, subProject := range changedSubprojects {
		dfs(subProject)
	}

	return result
}