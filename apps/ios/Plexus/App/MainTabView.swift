import SwiftUI

struct MainTabView: View {
    @State private var selectedOrgSlug: String?
    @State private var selectedProjectKey: String?

    var body: some View {
        TabView {
            NavigationStack {
                OrgsListView(
                    selectedOrgSlug: $selectedOrgSlug,
                    selectedProjectKey: $selectedProjectKey
                )
            }
            .tabItem {
                Label("Projects", systemImage: "square.grid.2x2")
            }

            NavigationStack {
                NotificationsView()
            }
            .tabItem {
                Label("Notifications", systemImage: "bell")
            }

            NavigationStack {
                ProfileView()
            }
            .tabItem {
                Label("Profile", systemImage: "person.circle")
            }
        }
    }
}

// MARK: - Orgs list

struct OrgsListView: View {
    @Binding var selectedOrgSlug: String?
    @Binding var selectedProjectKey: String?
    @State private var orgs: [Organization] = []
    @State private var isLoading = true
    @State private var showCreateSheet = false
    private let api = APIClient.shared

    var body: some View {
        Group {
            if isLoading {
                ProgressView()
            } else {
                List(orgs) { org in
                    NavigationLink(org.name) {
                        ProjectsListView(orgSlug: org.slug)
                    }
                    .subtitle(org.plan.capitalized)
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

extension NavigationLink where Label == Text {
    func subtitle(_ text: String) -> some View {
        HStack {
            self
            Spacer()
            Text(text)
                .font(.caption)
                .foregroundStyle(.secondary)
        }
    }
}

// MARK: - Projects list

struct ProjectsListView: View {
    let orgSlug: String
    @State private var projects: [Project] = []
    @State private var isLoading = true
    @State private var showCreateSheet = false
    private let api = APIClient.shared

    var body: some View {
        Group {
            if isLoading {
                ProgressView()
            } else {
                List(projects) { project in
                    Section {
                        NavigationLink("Board") {
                            BoardView(orgSlug: orgSlug, projectKey: project.key)
                        }
                        NavigationLink("Backlog") {
                            BacklogView(orgSlug: orgSlug, projectKey: project.key)
                        }
                    } header: {
                        Text("\(project.key) — \(project.name)")
                    }
                }
            }
        }
        .navigationTitle(orgSlug)
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
        defer { isLoading = false }
        do {
            let res: ListResponse<Project> = try await api.request("orgs/\(orgSlug)/projects")
            projects = res.items
        } catch {}
    }
}

// MARK: - Notifications

@MainActor
final class NotificationsViewModel: ObservableObject {
    @Published var notifications: [PlexusNotification] = []
    @Published var isLoading = false
    @Published var isMarkingRead = false

    private let api = APIClient.shared

    func load() async {
        isLoading = notifications.isEmpty
        defer { isLoading = false }
        do {
            let res: ListResponse<PlexusNotification> = try await api.request("notifications")
            notifications = res.items
        } catch {}
    }

    func markRead(id: String) async {
        guard let index = notifications.firstIndex(where: { $0.id == id }),
              !notifications[index].read else { return }
        isMarkingRead = true
        defer { isMarkingRead = false }
        do {
            try await api.requestEmpty("notifications/\(id)/read", method: "POST")
            await load()
        } catch {}
    }

    func markAllRead() async {
        isMarkingRead = true
        defer { isMarkingRead = false }
        do {
            try await api.requestEmpty("notifications/read-all", method: "POST")
            await load()
        } catch {}
    }
}

struct NotificationsView: View {
    @StateObject private var vm = NotificationsViewModel()

    var body: some View {
        Group {
            if vm.isLoading && vm.notifications.isEmpty {
                ProgressView()
            } else if vm.notifications.isEmpty {
                ContentUnavailableView("No notifications", systemImage: "bell.slash")
            } else {
                List(vm.notifications) { n in
                    Button {
                        Task { await vm.markRead(id: n.id) }
                    } label: {
                        VStack(alignment: .leading, spacing: 4) {
                            Text(n.title).font(.subheadline.weight(.medium))
                            if let body = n.body {
                                Text(body).font(.caption).foregroundStyle(.secondary)
                            }
                            Text(n.createdAt, style: .relative)
                                .font(.caption2).foregroundStyle(.tertiary)
                        }
                        .opacity(n.read ? 0.6 : 1)
                    }
                    .buttonStyle(.plain)
                }
            }
        }
        .navigationTitle("Notifications")
        .toolbar {
            if vm.notifications.contains(where: { !$0.read }) {
                ToolbarItem(placement: .primaryAction) {
                    Button("Mark all read") {
                        Task { await vm.markAllRead() }
                    }
                    .disabled(vm.isMarkingRead)
                }
            }
        }
        .task { await vm.load() }
        .refreshable { await vm.load() }
    }
}

// MARK: - Profile

struct ProfileView: View {
    @EnvironmentObject var authStore: AuthStore

    var body: some View {
        List {
            if let user = authStore.currentUser {
                Section {
                    HStack(spacing: 12) {
                        Circle()
                            .fill(.secondary)
                            .frame(width: 48, height: 48)
                            .overlay(
                                Text(user.displayName.prefix(1).uppercased())
                                    .font(.title2.bold())
                                    .foregroundStyle(.background)
                            )
                        VStack(alignment: .leading) {
                            Text(user.displayName).font(.headline)
                            Text(user.email).font(.subheadline).foregroundStyle(.secondary)
                        }
                    }
                    .padding(.vertical, 4)
                }
            }

            Section {
                Button("Sign out", role: .destructive) {
                    authStore.logout()
                }
            }
        }
        .navigationTitle("Profile")
    }
}
