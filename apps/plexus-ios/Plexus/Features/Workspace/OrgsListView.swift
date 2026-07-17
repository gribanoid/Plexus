import SwiftUI

struct OrgsListView: View {
    @State private var orgs: [Organization] = []
    @State private var isLoading = true
    @State private var showCreateSheet = false
    private let api = APIClient.shared

    var body: some View {
        Group {
            if isLoading {
                ProgressView()
            } else if orgs.isEmpty {
                ContentUnavailableView(
                    "No workspaces",
                    systemImage: "square.grid.2x2",
                    description: Text("Create your first workspace to get started.")
                )
            } else {
                List(orgs) { org in
                    NavigationLink {
                        ProjectsListView(orgSlug: org.slug, orgName: org.name)
                    } label: {
                        Text(org.name).font(.body.weight(.medium))
                    }
                }
            }
        }
        .navigationTitle("Workspaces")
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
            CreateWorkspaceSheet {
                Task { await loadOrgs() }
            }
        }
        .task { await loadOrgs() }
        .refreshable { await loadOrgs() }
    }

    private func loadOrgs() async {
        isLoading = orgs.isEmpty
        defer { isLoading = false }
        do {
            let res: ListResponse<Organization> = try await api.request("orgs")
            orgs = res.items
        } catch {}
    }
}
