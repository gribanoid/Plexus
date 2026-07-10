import SwiftUI

struct CreateIssueSheet: View {
    let orgSlug: String
    let projectKey: String
    var sprintId: String?
    var onCreated: () -> Void

    @Environment(\.dismiss) private var dismiss
    @State private var title = ""
    @State private var selectedTypeId = ""
    @State private var priority: Priority = .medium
    @State private var issueTypes: [IssueType] = []
    @State private var isLoadingTypes = true
    @State private var isSubmitting = false
    @State private var errorMessage: String?

    private let api = APIClient.shared

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("What needs to be done?", text: $title)
                }

                Section {
                    if isLoadingTypes {
                        ProgressView()
                    } else if issueTypes.isEmpty {
                        Text("No issue types available")
                            .foregroundStyle(.secondary)
                    } else {
                        Picker("Type", selection: $selectedTypeId) {
                            ForEach(issueTypes) { type in
                                Text(type.name).tag(type.id)
                            }
                        }
                        Picker("Priority", selection: $priority) {
                            ForEach(Priority.allCases, id: \.self) { p in
                                Text(p.displayName).tag(p)
                            }
                        }
                    }
                }

                if let errorMessage {
                    Section {
                        Text(errorMessage)
                            .foregroundStyle(.red)
                            .font(.caption)
                    }
                }
            }
            .navigationTitle("Create issue")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(isSubmitting ? "Creating…" : "Create") {
                        Task { await create() }
                    }
                    .disabled(
                        isSubmitting
                            || title.trimmingCharacters(in: .whitespaces).isEmpty
                            || selectedTypeId.isEmpty
                    )
                }
            }
            .task { await loadTypes() }
        }
    }

    private func loadTypes() async {
        isLoadingTypes = true
        defer { isLoadingTypes = false }
        do {
            let res: ListResponse<IssueType> = try await api.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/issue-types"
            )
            issueTypes = res.items
            if let first = issueTypes.first {
                selectedTypeId = first.id
            }
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    private func create() async {
        isSubmitting = true
        errorMessage = nil
        defer { isSubmitting = false }

        do {
            struct Body: Encodable {
                let title: String
                let typeId: String
                let priority: Priority
                let sprintId: String?
            }
            let _: CreateIssueResponse = try await api.request(
                "orgs/\(orgSlug)/projects/\(projectKey)/issues",
                method: "POST",
                body: Body(
                    title: title.trimmingCharacters(in: .whitespaces),
                    typeId: selectedTypeId,
                    priority: priority,
                    sprintId: sprintId
                )
            )
            onCreated()
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

struct CreateIssueResponse: Decodable {
    let id: String
    let number: Int
}
