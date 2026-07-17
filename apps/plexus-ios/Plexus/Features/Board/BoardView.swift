import SwiftUI

@MainActor
final class BoardViewModel: ObservableObject {
    @Published var statuses: [Status] = []
    @Published var issues: [Issue] = []
    @Published var issueTypes: [IssueType] = []
    @Published var members: [OrgMember] = []
    @Published var isLoading = false
    @Published var error: String?

    private let api = APIClient.shared

    func load(orgSlug: String, projectKey: String) async {
        isLoading = true
        defer { isLoading = false }
        do {
            async let statusRes: ListResponse<Status> = api.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/statuses"
            )
            async let issueRes: ListResponse<Issue> = api.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/issues"
            )
            async let typesRes: ListResponse<IssueType> = api.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/issue-types"
            )
            async let membersRes: ListResponse<OrgMember> = api.request(
                "orgs/\(orgSlug)/members"
            )
            let (s, i, t, m) = try await (statusRes, issueRes, typesRes, membersRes)
            statuses = s.items.sorted { $0.position < $1.position }
            issues = i.items
            issueTypes = t.items
            members = m.items
        } catch {
            self.error = error.localizedDescription
        }
    }

    func issuesFor(statusId: String) -> [Issue] {
        issues.filter { $0.statusId == statusId }.sorted { $0.position < $1.position }
    }

    func issueType(for issue: Issue) -> IssueType? {
        issueTypes.first { $0.id == issue.typeId }
    }

    func assignee(for issue: Issue) -> OrgMember? {
        guard let assigneeId = issue.assigneeId else { return nil }
        return members.first { $0.id == assigneeId }
    }

    func moveIssue(
        orgSlug: String,
        projectKey: String,
        issueNumber: Int,
        statusId: String
    ) async {
        do {
            struct Body: Encodable {
                let statusId: String
            }
            try await api.requestEmpty(
                "orgs/\(orgSlug)/projects/\(projectKey)/issues/\(issueNumber)/move",
                method: "POST",
                body: Body(statusId: statusId)
            )
            await load(orgSlug: orgSlug, projectKey: projectKey)
        } catch {
            self.error = error.localizedDescription
        }
    }
}

struct BoardView: View {
    let orgSlug: String
    let projectKey: String
    var embedded: Bool = false

    @StateObject private var vm = BoardViewModel()
    @State private var showCreateSheet = false
    @State private var createStatusId: String?

    var body: some View {
        Group {
            if vm.isLoading && vm.statuses.isEmpty {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else {
                ScrollView(.horizontal, showsIndicators: false) {
                    HStack(alignment: .top, spacing: 12) {
                        ForEach(vm.statuses) { status in
                            BoardColumn(
                                status: status,
                                issues: vm.issuesFor(statusId: status.id),
                                allStatuses: vm.statuses,
                                orgSlug: orgSlug,
                                projectKey: projectKey,
                                issueType: { vm.issueType(for: $0) },
                                assignee: { vm.assignee(for: $0) },
                                onMoveIssue: { issueNumber, newStatusId in
                                    Task {
                                        await vm.moveIssue(
                                            orgSlug: orgSlug,
                                            projectKey: projectKey,
                                            issueNumber: issueNumber,
                                            statusId: newStatusId
                                        )
                                    }
                                },
                                onCreate: {
                                    createStatusId = status.id
                                    showCreateSheet = true
                                }
                            )
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                }
            }
        }
        .navigationTitle(embedded ? "" : projectKey)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            if !embedded {
                ToolbarItem(placement: .primaryAction) {
                    Button {
                        createStatusId = nil
                        showCreateSheet = true
                    } label: {
                        Image(systemName: "plus")
                    }
                }
            }
        }
        .sheet(isPresented: $showCreateSheet) {
            CreateIssueSheet(
                orgSlug: orgSlug,
                projectKey: projectKey,
                statusId: createStatusId
            ) {
                Task { await vm.load(orgSlug: orgSlug, projectKey: projectKey) }
            }
        }
        .task { await vm.load(orgSlug: orgSlug, projectKey: projectKey) }
        .refreshable { await vm.load(orgSlug: orgSlug, projectKey: projectKey) }
    }
}

struct BoardColumn: View {
    let status: Status
    let issues: [Issue]
    let allStatuses: [Status]
    let orgSlug: String
    let projectKey: String
    let issueType: (Issue) -> IssueType?
    let assignee: (Issue) -> OrgMember?
    let onMoveIssue: (Int, String) -> Void
    let onCreate: () -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 0) {
            HStack(spacing: 8) {
                Text(status.name.uppercased())
                    .font(.caption.weight(.bold))
                    .foregroundStyle(.secondary)
                Text("\(issues.count)")
                    .font(.caption)
                    .foregroundStyle(.tertiary)
                Spacer()
                Image(systemName: "ellipsis")
                    .font(.caption)
                    .foregroundStyle(.tertiary)
            }
            .padding(.horizontal, 12)
            .padding(.vertical, 10)

            ScrollView {
                LazyVStack(spacing: 8) {
                    ForEach(issues) { issue in
                        NavigationLink {
                            IssueDetailView(
                                orgSlug: orgSlug,
                                projectKey: projectKey,
                                issueNumber: issue.number
                            )
                        } label: {
                            IssueCard(
                                issue: issue,
                                projectKey: projectKey,
                                issueType: issueType(issue),
                                assignee: assignee(issue)
                            )
                        }
                        .buttonStyle(.plain)
                        .contextMenu {
                            ForEach(allStatuses.filter { $0.id != issue.statusId }) { target in
                                Button("Move to \(target.name)") {
                                    onMoveIssue(issue.number, target.id)
                                }
                            }
                        }
                    }

                    Button(action: onCreate) {
                        Label("Create", systemImage: "plus")
                            .font(.subheadline)
                            .frame(maxWidth: .infinity, alignment: .leading)
                            .padding(.vertical, 8)
                            .padding(.horizontal, 4)
                    }
                    .buttonStyle(.plain)
                    .foregroundStyle(Color(hex: "#0052CC"))
                }
                .padding(.horizontal, 8)
                .padding(.bottom, 12)
            }
        }
        .frame(width: 280)
        .background(Color(.secondarySystemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 10))
    }
}

struct IssueCard: View {
    let issue: Issue
    let projectKey: String
    let issueType: IssueType?
    let assignee: OrgMember?

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            Text(issue.title)
                .font(.subheadline)
                .foregroundStyle(.primary)
                .lineLimit(3)
                .multilineTextAlignment(.leading)

            HStack(spacing: 8) {
                RoundedRectangle(cornerRadius: 3)
                    .fill(Color(hex: issueType?.color ?? "#0052CC"))
                    .frame(width: 14, height: 14)
                Text("\(projectKey)-\(issue.number)")
                    .font(.caption2)
                    .foregroundStyle(.secondary)
                    .monospacedDigit()
                Spacer()
                if let assignee {
                    Circle()
                        .fill(Color(hex: "#DEEBFF"))
                        .frame(width: 22, height: 22)
                        .overlay(
                            Text(assignee.displayName.prefix(1).uppercased())
                                .font(.caption2.bold())
                                .foregroundStyle(Color(hex: "#0052CC"))
                        )
                }
            }
        }
        .padding(12)
        .frame(maxWidth: .infinity, alignment: .leading)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 8))
        .shadow(color: .black.opacity(0.05), radius: 2, y: 1)
    }
}
