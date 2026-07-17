import SwiftUI

enum ProjectHubTab: String, CaseIterable, Identifiable {
    case board = "Board"
    case backlog = "Backlog"
    case roadmap = "Roadmap"
    case reports = "Reports"
    case settings = "Settings"

    var id: String { rawValue }
}

struct ProjectHubView: View {
    let orgSlug: String
    let projectKey: String
    let projectName: String

    @State private var selectedTab: ProjectHubTab = .board
    @State private var statuses: [Status] = []

    var body: some View {
        VStack(spacing: 0) {
            ScrollView(.horizontal, showsIndicators: false) {
                HStack(spacing: 0) {
                    ForEach(ProjectHubTab.allCases) { tab in
                        Button {
                            selectedTab = tab
                        } label: {
                            VStack(spacing: 8) {
                                Text(tab.rawValue)
                                    .font(.subheadline.weight(selectedTab == tab ? .semibold : .regular))
                                    .foregroundStyle(selectedTab == tab ? Color(hex: "#0052CC") : .secondary)
                                    .padding(.horizontal, 14)
                                    .padding(.top, 10)
                                Rectangle()
                                    .fill(selectedTab == tab ? Color(hex: "#0052CC") : .clear)
                                    .frame(height: 3)
                            }
                        }
                        .buttonStyle(.plain)
                    }
                }
                .padding(.horizontal, 8)
            }
            .background(Color(.systemBackground))

            Divider()

            Group {
                switch selectedTab {
                case .board:
                    BoardView(orgSlug: orgSlug, projectKey: projectKey, embedded: true)
                case .backlog:
                    BacklogView(orgSlug: orgSlug, projectKey: projectKey, embedded: true)
                case .roadmap:
                    ComingSoonView(feature: "Roadmap")
                case .reports:
                    ComingSoonView(feature: "Reports")
                case .settings:
                    ProjectSettingsView(
                        orgSlug: orgSlug,
                        projectKey: projectKey,
                        statuses: statuses,
                        onStatusesChanged: { await loadStatuses() }
                    )
                }
            }
            .frame(maxWidth: .infinity, maxHeight: .infinity)
        }
        .navigationTitle(projectName)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button {} label: {
                    Image(systemName: "line.3.horizontal.decrease")
                }
            }
            ToolbarItem(placement: .topBarTrailing) {
                Button {} label: {
                    Image(systemName: "ellipsis")
                }
            }
        }
        .task { await loadStatuses() }
    }

    private func loadStatuses() async {
        do {
            let res: ListResponse<Status> = try await APIClient.shared.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/statuses"
            )
            statuses = res.items.sorted { $0.position < $1.position }
        } catch {}
    }
}

struct ComingSoonView: View {
    let feature: String

    var body: some View {
        ContentUnavailableView(
            feature,
            systemImage: "clock",
            description: Text("Coming soon")
        )
    }
}

struct ProjectSettingsView: View {
    let orgSlug: String
    let projectKey: String
    let statuses: [Status]
    var onStatusesChanged: () async -> Void

    @State private var showAddStatus = false
    @State private var errorMessage: String?
    @State private var isDeleting = false

    private let api = APIClient.shared

    var body: some View {
        List {
            Section {
                Text("Manage workflow columns for \(projectKey). Other settings are available on the web.")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }

            Section {
                if let errorMessage {
                    Text(errorMessage)
                        .font(.caption)
                        .foregroundStyle(.red)
                }

                if statuses.isEmpty {
                    Text("No statuses yet").foregroundStyle(.secondary)
                } else {
                    ForEach(statuses) { status in
                        HStack {
                            Circle()
                                .fill(Color(hex: status.color))
                                .frame(width: 8, height: 8)
                            Text(status.name)
                            Spacer()
                            Text(status.category.displayName)
                                .font(.caption)
                                .foregroundStyle(.secondary)
                        }
                        .swipeActions(edge: .trailing, allowsFullSwipe: true) {
                            Button(role: .destructive) {
                                Task { await deleteStatus(status.id) }
                            } label: {
                                Label("Delete", systemImage: "trash")
                            }
                            .disabled(isDeleting)
                        }
                    }
                }
            } header: {
                Text("Workflow")
            } footer: {
                Button {
                    showAddStatus = true
                } label: {
                    Label("Add status", systemImage: "plus.circle.fill")
                }
                .font(.subheadline.weight(.medium))
                .foregroundStyle(Color(hex: "#0052CC"))
            }
        }
        .sheet(isPresented: $showAddStatus) {
            AddStatusSheet(orgSlug: orgSlug, projectKey: projectKey) {
                await onStatusesChanged()
            }
        }
    }

    private func deleteStatus(_ id: String) async {
        isDeleting = true
        errorMessage = nil
        defer { isDeleting = false }
        do {
            try await api.requestEmpty(
                "orgs/\(orgSlug)/projects/\(projectKey)/statuses/\(id)",
                method: "DELETE"
            )
            await onStatusesChanged()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct AddStatusSheet: View {
    let orgSlug: String
    let projectKey: String
    var onCreated: () async -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var name = ""
    @State private var color = "#4C9AFF"
    @State private var category: StatusCategory = .todo
    @State private var isSubmitting = false
    @State private var errorMessage: String?

    private let api = APIClient.shared
    private let colorPresets = ["#4C9AFF", "#6B7280", "#22C55E", "#F97316", "#EF4444", "#8B5CF6", "#EAB308"]

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Name", text: $name)
                        .autocorrectionDisabled()
                }

                Section("Color") {
                    HStack(spacing: 12) {
                        ForEach(colorPresets, id: \.self) { hex in
                            Button {
                                color = hex
                            } label: {
                                Circle()
                                    .fill(Color(hex: hex))
                                    .frame(width: 28, height: 28)
                                    .overlay {
                                        if color == hex {
                                            Circle().strokeBorder(.primary, lineWidth: 2)
                                        }
                                    }
                            }
                            .buttonStyle(.plain)
                        }
                    }
                    .padding(.vertical, 4)
                }

                Section("Category") {
                    Picker("Category", selection: $category) {
                        ForEach(StatusCategory.allCases) { cat in
                            Text(cat.displayName).tag(cat)
                        }
                    }
                    .pickerStyle(.segmented)
                }

                if let errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .font(.caption)
                    }
                }
            }
            .navigationTitle("Add status")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(isSubmitting ? "Adding…" : "Add") {
                        Task { await create() }
                    }
                    .disabled(isSubmitting || name.trimmingCharacters(in: .whitespaces).isEmpty)
                }
            }
        }
    }

    private func create() async {
        isSubmitting = true
        errorMessage = nil
        defer { isSubmitting = false }

        do {
            struct Body: Encodable {
                let name: String
                let color: String
                let category: String
            }
            let _: CreatedID = try await api.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/statuses",
                method: "POST",
                body: Body(
                    name: name.trimmingCharacters(in: .whitespaces),
                    color: color,
                    category: category.rawValue
                )
            )
            await onCreated()
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
