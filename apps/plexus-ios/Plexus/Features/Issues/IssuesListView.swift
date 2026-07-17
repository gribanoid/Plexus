import SwiftUI

struct IssueListItem: Identifiable {
    var id: String { "\(projectKey)-\(issue.id)" }
    let issue: Issue
    let orgSlug: String
    let projectKey: String
    let projectName: String
}

@MainActor
final class IssuesListViewModel: ObservableObject {
    @Published var orgs: [Organization] = []
    @Published var selectedOrgSlug: String?
    @Published var items: [IssueListItem] = []
    @Published var isLoading = false
    @Published var error: String?

    private let api = APIClient.shared

    func load(orgSlug: String? = nil) async {
        isLoading = true
        error = nil
        defer { isLoading = false }
        do {
            let orgRes: ListResponse<Organization> = try await api.request("orgs")
            orgs = orgRes.items
            let slug = orgSlug ?? selectedOrgSlug ?? orgs.first?.slug
            selectedOrgSlug = slug
            guard let slug else {
                items = []
                return
            }
            let projectRes: ListResponse<Project> = try await api.request("orgs/\(slug)/projects")
            var collected: [IssueListItem] = []
            var issueErrors: [String] = []
            for project in projectRes.items {
                do {
                    let issueRes: ListResponse<Issue> = try await api.request(
                        "orgs/\(slug)/projects/\(project.key)/issues"
                    )
                    collected.append(contentsOf: issueRes.items.map {
                        IssueListItem(
                            issue: $0,
                            orgSlug: slug,
                            projectKey: project.key,
                            projectName: project.name
                        )
                    })
                } catch {
                    issueErrors.append("\(project.key): \(error.localizedDescription)")
                }
            }
            items = collected.sorted { $0.issue.updatedAt > $1.issue.updatedAt }
            if items.isEmpty, let first = issueErrors.first {
                self.error = first
            }
        } catch {
            self.error = error.localizedDescription
        }
    }
}

struct IssuesListView: View {
    @StateObject private var vm = IssuesListViewModel()

    var body: some View {
        Group {
            if vm.isLoading && vm.items.isEmpty {
                ProgressView()
            } else if let error = vm.error {
                ContentUnavailableView("Error", systemImage: "exclamationmark.triangle", description: Text(error))
            } else if vm.items.isEmpty {
                ContentUnavailableView("No issues", systemImage: "checklist")
            } else {
                List {
                    if !vm.orgs.isEmpty {
                        Section {
                            Picker("Workspace", selection: Binding(
                                get: { vm.selectedOrgSlug ?? "" },
                                set: { newValue in
                                    Task { await vm.load(orgSlug: newValue) }
                                }
                            )) {
                                ForEach(vm.orgs) { org in
                                    Text(org.name).tag(org.slug)
                                }
                            }
                        }
                    }

                    Section {
                        ForEach(vm.items) { item in
                            NavigationLink {
                                IssueDetailView(
                                    orgSlug: item.orgSlug,
                                    projectKey: item.projectKey,
                                    issueNumber: item.issue.number
                                )
                            } label: {
                                VStack(alignment: .leading, spacing: 4) {
                                    Text(item.issue.title)
                                        .font(.subheadline.weight(.medium))
                                        .lineLimit(2)
                                    Text("\(item.projectKey)-\(item.issue.number) · \(item.projectName)")
                                        .font(.caption)
                                        .foregroundStyle(.secondary)
                                }
                            }
                        }
                    }
                }
            }
        }
        .navigationTitle("Issues")
        .task { await vm.load() }
        .refreshable { await vm.load() }
    }
}
