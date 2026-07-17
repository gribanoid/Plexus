import SwiftUI

@MainActor
final class BacklogViewModel: ObservableObject {
    @Published var sprints: [Sprint] = []
    @Published var issues: [Issue] = []
    @Published var statuses: [Status] = []
    @Published var isLoading = false
    @Published var actingSprintId: String?

    private let api = APIClient.shared

    func load(orgSlug: String, projectKey: String) async {
        isLoading = true
        defer { isLoading = false }
        do {
            async let sprintRes: ListResponse<Sprint> = api.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/sprints"
            )
            async let issueRes: ListResponse<Issue> = api.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/issues"
            )
            async let statusRes: ListResponse<Status> = api.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/statuses"
            )
            let (sp, is_, st) = try await (sprintRes, issueRes, statusRes)
            sprints = sp.items
            issues = is_.items
            statuses = st.items
        } catch {}
    }

    func status(for issue: Issue) -> Status? {
        statuses.first { $0.id == issue.statusId }
    }

    func startSprint(orgSlug: String, projectKey: String, sprintId: String) async {
        actingSprintId = sprintId
        defer { actingSprintId = nil }
        do {
            try await api.requestEmpty(
                "orgs/\(orgSlug)/projects/\(projectKey)/sprints/\(sprintId)/start",
                method: "POST"
            )
            await load(orgSlug: orgSlug, projectKey: projectKey)
        } catch {}
    }

    func completeSprint(orgSlug: String, projectKey: String, sprintId: String) async {
        actingSprintId = sprintId
        defer { actingSprintId = nil }
        do {
            try await api.requestEmpty(
                "orgs/\(orgSlug)/projects/\(projectKey)/sprints/\(sprintId)/complete",
                method: "POST"
            )
            await load(orgSlug: orgSlug, projectKey: projectKey)
        } catch {}
    }
}

struct BacklogView: View {
    let orgSlug: String
    let projectKey: String
    var embedded: Bool = false
    @StateObject private var vm = BacklogViewModel()
    @State private var expandedSprints: Set<String> = []
    @State private var showCreateSheet = false
    @State private var createSprintId: String?

    var activeSprint: Sprint? { vm.sprints.first { $0.state == .active } }
    var futureSprints: [Sprint] { vm.sprints.filter { $0.state == .future } }

    var body: some View {
        List {
            if let sprint = activeSprint {
                SprintSection(
                    sprint: sprint,
                    issues: vm.issues.filter { $0.sprintId == sprint.id },
                    orgSlug: orgSlug,
                    projectKey: projectKey,
                    vm: vm,
                    isExpanded: !expandedSprints.contains(sprint.id),
                    onToggle: { toggle(sprint.id) },
                    sprintActionLabel: "Complete sprint",
                    onSprintAction: {
                        Task { await vm.completeSprint(orgSlug: orgSlug, projectKey: projectKey, sprintId: sprint.id) }
                    }
                )
            }

            ForEach(futureSprints) { sprint in
                SprintSection(
                    sprint: sprint,
                    issues: vm.issues.filter { $0.sprintId == sprint.id },
                    orgSlug: orgSlug,
                    projectKey: projectKey,
                    vm: vm,
                    isExpanded: !expandedSprints.contains(sprint.id),
                    onToggle: { toggle(sprint.id) },
                    sprintActionLabel: "Start sprint",
                    onSprintAction: {
                        Task { await vm.startSprint(orgSlug: orgSlug, projectKey: projectKey, sprintId: sprint.id) }
                    }
                )
            }

            Section {
                let backlog = vm.issues.filter { $0.sprintId == nil }
                ForEach(backlog) { issue in
                    NavigationLink {
                        IssueDetailView(
                            orgSlug: orgSlug,
                            projectKey: projectKey,
                            issueNumber: issue.number
                        )
                    } label: {
                        IssueRow(issue: issue, projectKey: projectKey, status: vm.status(for: issue))
                    }
                }
                if backlog.isEmpty {
                    Text("No issues in backlog")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                        .frame(maxWidth: .infinity, alignment: .center)
                        .padding(.vertical, 12)
                }
            } header: {
                Text("Backlog · \(vm.issues.filter { $0.sprintId == nil }.count)")
                    .font(.caption.weight(.semibold))
            }
        }
        .listStyle(.insetGrouped)
        .navigationTitle(embedded ? "" : "Backlog")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button {
                    createSprintId = nil
                    showCreateSheet = true
                } label: {
                    Image(systemName: "plus")
                }
            }
        }
        .sheet(isPresented: $showCreateSheet) {
            CreateIssueSheet(
                orgSlug: orgSlug,
                projectKey: projectKey,
                sprintId: createSprintId
            ) {
                Task { await vm.load(orgSlug: orgSlug, projectKey: projectKey) }
            }
        }
        .task { await vm.load(orgSlug: orgSlug, projectKey: projectKey) }
        .refreshable { await vm.load(orgSlug: orgSlug, projectKey: projectKey) }
    }

    func toggle(_ id: String) {
        if expandedSprints.contains(id) {
            expandedSprints.remove(id)
        } else {
            expandedSprints.insert(id)
        }
    }
}

struct SprintSection: View {
    let sprint: Sprint
    let issues: [Issue]
    let orgSlug: String
    let projectKey: String
    let vm: BacklogViewModel
    let isExpanded: Bool
    let onToggle: () -> Void
    var sprintActionLabel: String?
    var onSprintAction: (() -> Void)?

    var body: some View {
        Section(isExpanded: .constant(isExpanded)) {
            ForEach(issues) { issue in
                NavigationLink {
                    IssueDetailView(
                        orgSlug: orgSlug,
                        projectKey: projectKey,
                        issueNumber: issue.number
                    )
                } label: {
                    IssueRow(issue: issue, projectKey: projectKey, status: vm.status(for: issue))
                }
            }
        } header: {
            HStack {
                Button(action: onToggle) {
                    HStack {
                        Image(systemName: isExpanded ? "chevron.down" : "chevron.right")
                            .font(.caption2)
                            .foregroundStyle(.secondary)
                        Text(sprint.name)
                            .font(.caption.weight(.semibold))
                        if sprint.state == .active {
                            Text("ACTIVE")
                                .font(.caption2.weight(.bold))
                                .foregroundStyle(.green)
                        }
                        Spacer()
                        Text("\(issues.count)")
                            .font(.caption)
                            .foregroundStyle(.tertiary)
                    }
                }
                .buttonStyle(.plain)

                if let sprintActionLabel, let onSprintAction {
                    Button {
                        onSprintAction()
                    } label: {
                        if vm.actingSprintId == sprint.id {
                            ProgressView()
                                .controlSize(.small)
                        } else {
                            Text(sprintActionLabel)
                                .font(.caption.weight(.medium))
                        }
                    }
                    .buttonStyle(.bordered)
                    .controlSize(.small)
                    .disabled(vm.actingSprintId != nil)
                }
            }
        }
    }
}

struct IssueRow: View {
    let issue: Issue
    let projectKey: String
    let status: Status?

    var body: some View {
        HStack(spacing: 10) {
            PriorityDot(priority: issue.priority)

            VStack(alignment: .leading, spacing: 2) {
                Text(issue.title)
                    .font(.subheadline)
                    .lineLimit(2)
                HStack(spacing: 6) {
                    Text("\(projectKey)-\(issue.number)")
                        .font(.caption2)
                        .foregroundStyle(.tertiary)
                        .monospacedDigit()
                    if let status {
                        Text(status.name)
                            .font(.caption2)
                            .foregroundStyle(.secondary)
                    }
                }
            }

            Spacer()

            if let sp = issue.storyPoints {
                Text("\(Int(sp))")
                    .font(.caption2.weight(.medium))
                    .padding(.horizontal, 5)
                    .padding(.vertical, 2)
                    .background(Color(.secondarySystemBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 4))
            }
        }
        .padding(.vertical, 2)
    }
}
