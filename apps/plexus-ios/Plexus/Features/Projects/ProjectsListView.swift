import SwiftUI

struct ProjectsListView: View {
    let orgSlug: String
    var orgName: String? = nil
    @State private var projects: [Project] = []
    @State private var isLoading = true
    @State private var error: String?
    @State private var showCreateSheet = false
    private let api = APIClient.shared

    var body: some View {
        Group {
            if isLoading {
                ProgressView()
            } else if let error {
                ContentUnavailableView("Error", systemImage: "exclamationmark.triangle", description: Text(error))
            } else if projects.isEmpty {
                ContentUnavailableView(
                    "No projects",
                    systemImage: "folder",
                    description: Text("Create a project to open the board.")
                )
            } else {
                List(projects) { project in
                    NavigationLink {
                        ProjectHubView(
                            orgSlug: orgSlug,
                            projectKey: project.key,
                            projectName: project.name
                        )
                    } label: {
                        HStack(spacing: 12) {
                            Text(project.key)
                                .font(.caption.weight(.bold))
                                .foregroundStyle(Color(hex: "#0052CC"))
                                .frame(width: 44, height: 36)
                                .background(Color(hex: "#DEEBFF"))
                                .clipShape(RoundedRectangle(cornerRadius: 6))
                            VStack(alignment: .leading, spacing: 2) {
                                Text(project.name).font(.body.weight(.medium))
                                Text(project.description ?? "No description")
                                    .font(.caption)
                                    .foregroundStyle(.secondary)
                                    .lineLimit(1)
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle(orgName ?? orgSlug)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    showCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .sheet(isPresented: $showCreateSheet) {
            CreateProjectSheet(orgSlug: orgSlug) {
                Task { await loadProjects() }
            }
        }
        .task { await loadProjects() }
        .refreshable { await loadProjects() }
    }

    private func loadProjects() async {
        isLoading = projects.isEmpty
        error = nil
        defer { isLoading = false }
        do {
            let res: ListResponse<Project> = try await api.request("orgs/\(orgSlug)/projects")
            projects = res.items
        } catch {
            self.error = error.localizedDescription
        }
    }
}
