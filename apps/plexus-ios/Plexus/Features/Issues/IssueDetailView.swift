import SwiftUI

@MainActor
final class IssueDetailViewModel: ObservableObject {
    @Published var issue: Issue?
    @Published var statuses: [Status] = []
    @Published var issueTypes: [IssueType] = []
    @Published var comments: [Comment] = []
    @Published var history: [IssueHistory] = []
    @Published var members: [OrgMember] = []
    @Published var isLoading = false
    @Published var isSubmittingComment = false
    @Published var isSaving = false
    @Published var error: String?

    private let api = APIClient.shared
    private let orgSlug: String
    private let issuePath: String
    private let projectBasePath: String

    init(orgSlug: String, projectKey: String, issueNumber: Int) {
        self.orgSlug = orgSlug
        projectBasePath = "orgs/\(orgSlug)/projects/\(projectKey)"
        issuePath = "\(projectBasePath)/issues/\(issueNumber)"
    }

    func load() async {
        isLoading = true
        defer { isLoading = false }
        do {
            async let issueReq: Issue = api.request(issuePath)
            async let statusReq: ListResponse<Status> = api.request("\(projectBasePath)/statuses")
            async let typesReq: ListResponse<IssueType> = api.request("\(projectBasePath)/issue-types")
            async let commentsReq: ListResponse<Comment> = api.request("\(issuePath)/comments")
            async let historyReq: ListResponse<IssueHistory> = api.request("\(issuePath)/history")

            let (i, s, t, c, h) = try await (issueReq, statusReq, typesReq, commentsReq, historyReq)
            issue = i
            statuses = s.items
            issueTypes = t.items
            comments = c.items
            history = h.items
        } catch {
            self.error = error.localizedDescription
        }
    }

    func addComment(body: String) async {
        isSubmittingComment = true
        defer { isSubmittingComment = false }
        do {
            struct Body: Encodable { let body: String }
            struct CreateResponse: Decodable { let id: String }
            let _: CreateResponse = try await api.request(
                "\(issuePath)/comments",
                method: "POST",
                body: Body(body: body)
            )
            let res: ListResponse<Comment> = try await api.request("\(issuePath)/comments")
            comments = res.items
        } catch {
            self.error = error.localizedDescription
        }
    }

    func status(for issue: Issue) -> Status? {
        statuses.first { $0.id == issue.statusId }
    }

    func issueType(for issue: Issue) -> IssueType? {
        issueTypes.first { $0.id == issue.typeId }
    }

    func memberName(for issue: Issue) -> String? {
        guard let assigneeId = issue.assigneeId else { return nil }
        return members.first { $0.id == assigneeId }?.displayName
    }

    func loadMembers() async {
        do {
            let res: ListResponse<OrgMember> = try await api.request("orgs/\(orgSlug)/members")
            members = res.items
        } catch {
            self.error = error.localizedDescription
        }
    }

    func updateIssue(title: String, statusId: String, assigneeId: String?) async -> Bool {
        isSaving = true
        defer { isSaving = false }
        do {
            struct Body: Encodable {
                let title: String
                let statusId: String
                let assigneeId: String?
            }
            try await api.requestEmpty(
                issuePath,
                method: "PATCH",
                body: Body(title: title, statusId: statusId, assigneeId: assigneeId)
            )
            await load()
            return true
        } catch {
            self.error = error.localizedDescription
            return false
        }
    }
}

struct IssueDetailView: View {
    let orgSlug: String
    let projectKey: String
    let issueNumber: Int

    @StateObject private var vm: IssueDetailViewModel
    @State private var commentBody = ""
    @State private var showEditSheet = false

    init(orgSlug: String, projectKey: String, issueNumber: Int) {
        self.orgSlug = orgSlug
        self.projectKey = projectKey
        self.issueNumber = issueNumber
        _vm = StateObject(wrappedValue: IssueDetailViewModel(
            orgSlug: orgSlug,
            projectKey: projectKey,
            issueNumber: issueNumber
        ))
    }

    var body: some View {
        Group {
            if vm.isLoading && vm.issue == nil {
                ProgressView()
                    .frame(maxWidth: .infinity, maxHeight: .infinity)
            } else if let issue = vm.issue {
                ScrollView {
                    VStack(alignment: .leading, spacing: 24) {
                        header(issue: issue)
                        detailsSection(issue: issue)
                        descriptionSection(issue: issue)
                        commentsSection
                        historySection
                    }
                    .padding()
                }
            } else {
                ContentUnavailableView("Issue not found", systemImage: "exclamationmark.triangle")
            }
        }
        .navigationTitle("\(projectKey)-\(issueNumber)")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button("Edit") { showEditSheet = true }
                    .disabled(vm.issue == nil)
            }
        }
        .sheet(isPresented: $showEditSheet) {
            if let issue = vm.issue {
                EditIssueSheet(
                    issue: issue,
                    statuses: vm.statuses,
                    members: vm.members
                ) { title, statusId, assigneeId in
                    await vm.updateIssue(title: title, statusId: statusId, assigneeId: assigneeId)
                }
            }
        }
        .task {
            await vm.load()
            await vm.loadMembers()
        }
        .refreshable {
            await vm.load()
            await vm.loadMembers()
        }
    }

    @ViewBuilder
    private func header(issue: Issue) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(issue.title)
                .font(.title2.bold())

            HStack(spacing: 8) {
                PriorityDot(priority: issue.priority)
                Text(issue.priority.displayName)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)

                if let status = vm.status(for: issue) {
                    StatusBadge(status: status)
                }
            }
        }
        .frame(maxWidth: .infinity, alignment: .leading)
    }

    @ViewBuilder
    private func detailsSection(issue: Issue) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Details")
                .font(.subheadline.weight(.semibold))

            VStack(spacing: 0) {
                if let type = vm.issueType(for: issue) {
                    DetailRow(label: "Type") {
                        HStack(spacing: 6) {
                            Circle()
                                .fill(Color(hex: type.color))
                                .frame(width: 8, height: 8)
                            Text(type.name)
                        }
                    }
                }
                DetailRow(label: "Priority") {
                    HStack(spacing: 6) {
                        PriorityDot(priority: issue.priority)
                        Text(issue.priority.displayName)
                    }
                }
                if let status = vm.status(for: issue) {
                    DetailRow(label: "Status") {
                        StatusBadge(status: status)
                    }
                }
                if let sp = issue.storyPoints {
                    DetailRow(label: "Story points") {
                        Text("\(Int(sp))")
                    }
                }
                DetailRow(label: "Assignee") {
                    Text(vm.memberName(for: issue) ?? "Unassigned")
                        .foregroundStyle(vm.memberName(for: issue) == nil ? .secondary : .primary)
                }
            }
            .background(Color(.secondarySystemGroupedBackground))
            .clipShape(RoundedRectangle(cornerRadius: 10))
        }
    }

    @ViewBuilder
    private func descriptionSection(issue: Issue) -> some View {
        VStack(alignment: .leading, spacing: 8) {
            Text("Description")
                .font(.subheadline.weight(.semibold))

            if let description = issue.description, !description.isEmpty {
                Text(description)
                    .font(.body)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding()
                    .background(Color(.secondarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
            } else {
                Text("No description")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .frame(maxWidth: .infinity, alignment: .leading)
                    .padding()
                    .background(Color(.secondarySystemGroupedBackground))
                    .clipShape(RoundedRectangle(cornerRadius: 10))
            }
        }
    }

    private var commentsSection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Comments (\(vm.comments.count))")
                .font(.subheadline.weight(.semibold))

            ForEach(vm.comments) { comment in
                VStack(alignment: .leading, spacing: 4) {
                    Text(comment.createdAt, style: .relative)
                        .font(.caption2)
                        .foregroundStyle(.tertiary)
                    Text(comment.body)
                        .font(.subheadline)
                        .padding(10)
                        .frame(maxWidth: .infinity, alignment: .leading)
                        .background(Color(.secondarySystemGroupedBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 8))
                }
            }

            HStack(alignment: .bottom, spacing: 8) {
                TextField("Add a comment…", text: $commentBody, axis: .vertical)
                    .lineLimit(1...4)
                    .textFieldStyle(.roundedBorder)

                Button {
                    let body = commentBody.trimmingCharacters(in: .whitespacesAndNewlines)
                    guard !body.isEmpty else { return }
                    Task {
                        await vm.addComment(body: body)
                        commentBody = ""
                    }
                } label: {
                    if vm.isSubmittingComment {
                        ProgressView()
                    } else {
                        Image(systemName: "paperplane.fill")
                    }
                }
                .disabled(commentBody.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty || vm.isSubmittingComment)
            }
        }
    }

    private var historySection: some View {
        VStack(alignment: .leading, spacing: 12) {
            Text("Activity")
                .font(.subheadline.weight(.semibold))

            if vm.history.isEmpty {
                Text("No history yet.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            } else {
                ForEach(vm.history) { entry in
                    VStack(alignment: .leading, spacing: 2) {
                        Text(historyText(for: entry))
                            .font(.subheadline)
                        Text(entry.createdAt, style: .relative)
                            .font(.caption2)
                            .foregroundStyle(.tertiary)
                    }
                    .padding(.vertical, 4)
                }
            }
        }
    }

    private func historyText(for entry: IssueHistory) -> String {
        var text = "\(entry.field) changed"
        if let newValue = entry.newValue {
            text += " to \(newValue)"
        }
        return text
    }
}

struct DetailRow<Content: View>: View {
    let label: String
    @ViewBuilder let content: Content

    var body: some View {
        HStack {
            Text(label)
                .font(.subheadline)
                .foregroundStyle(.secondary)
                .frame(width: 110, alignment: .leading)
            content
                .font(.subheadline)
            Spacer()
        }
        .padding(.horizontal, 12)
        .padding(.vertical, 10)
    }
}

struct StatusBadge: View {
    let status: Status

    var body: some View {
        HStack(spacing: 6) {
            Circle()
                .fill(Color(hex: status.color))
                .frame(width: 8, height: 8)
            Text(status.name)
                .font(.caption.weight(.medium))
        }
        .padding(.horizontal, 8)
        .padding(.vertical, 4)
        .background(Color(.tertiarySystemFill))
        .clipShape(Capsule())
    }
}
