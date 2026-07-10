import SwiftUI

@MainActor
final class BoardViewModel: ObservableObject {
    @Published var statuses: [Status] = []
    @Published var issues: [Issue] = []
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
            let (s, i) = try await (statusRes, issueRes)
            statuses = s.items.sorted { $0.position < $1.position }
            issues = i.items
        } catch {
            self.error = error.localizedDescription
        }
    }

    func issuesFor(statusId: String) -> [Issue] {
        issues.filter { $0.statusId == statusId }.sorted { $0.position < $1.position }
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
    @StateObject private var vm = BoardViewModel()
    @State private var showCreateSheet = false

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
                                onMoveIssue: { issueNumber, newStatusId in
                                    Task {
                                        await vm.moveIssue(
                                            orgSlug: orgSlug,
                                            projectKey: projectKey,
                                            issueNumber: issueNumber,
                                            statusId: newStatusId
                                        )
                                    }
                                }
                            )
                        }
                    }
                    .padding(.horizontal, 16)
                    .padding(.vertical, 12)
                }
            }
        }
        .navigationTitle(projectKey)
        .navigationBarTitleDisplayMode(.inline)
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
            CreateIssueSheet(orgSlug: orgSlug, projectKey: projectKey) {
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
    let onMoveIssue: (Int, String) -> Void

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            // Column header
            HStack {
                Circle()
                    .fill(Color(hex: status.color))
                    .frame(width: 8, height: 8)
                Text(status.name)
                    .font(.caption.weight(.semibold))
                    .foregroundStyle(.secondary)
                Spacer()
                Text("\(issues.count)")
                    .font(.caption)
                    .foregroundStyle(.tertiary)
            }
            .padding(.horizontal, 8)

            // Issue cards
            ForEach(issues) { issue in
                NavigationLink {
                    IssueDetailView(
                        orgSlug: orgSlug,
                        projectKey: projectKey,
                        issueNumber: issue.number
                    )
                } label: {
                    IssueCard(issue: issue, projectKey: projectKey)
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

            Spacer(minLength: 0)
        }
        .frame(width: 260)
        .padding(.vertical, 8)
        .background(Color(.systemGroupedBackground))
        .clipShape(RoundedRectangle(cornerRadius: 10))
    }
}

struct IssueCard: View {
    let issue: Issue
    let projectKey: String

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(issue.title)
                .font(.subheadline)
                .lineLimit(3)

            HStack(spacing: 6) {
                PriorityDot(priority: issue.priority)
                Text("\(projectKey)-\(issue.number)")
                    .font(.caption2)
                    .foregroundStyle(.tertiary)
                    .monospacedDigit()
                Spacer()
                if let sp = issue.storyPoints {
                    Text("\(Int(sp))")
                        .font(.caption2.weight(.medium))
                        .padding(.horizontal, 6)
                        .padding(.vertical, 2)
                        .background(Color(.secondarySystemBackground))
                        .clipShape(RoundedRectangle(cornerRadius: 4))
                }
            }
        }
        .padding(10)
        .background(Color(.systemBackground))
        .clipShape(RoundedRectangle(cornerRadius: 8))
        .shadow(color: .black.opacity(0.04), radius: 2, y: 1)
        .padding(.horizontal, 4)
    }
}

struct PriorityDot: View {
    let priority: Priority

    var body: some View {
        Circle()
            .fill(Color(hex: priority.color))
            .frame(width: 8, height: 8)
    }
}

// MARK: - Color hex extension

extension Color {
    init(hex: String) {
        let hex = hex.trimmingCharacters(in: CharacterSet.alphanumerics.inverted)
        var int: UInt64 = 0
        Scanner(string: hex).scanHexInt64(&int)
        let r = Double((int >> 16) & 0xFF) / 255
        let g = Double((int >> 8) & 0xFF) / 255
        let b = Double(int & 0xFF) / 255
        self.init(red: r, green: g, blue: b)
    }
}
