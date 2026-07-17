import SwiftUI

struct EditIssueSheet: View {
    let issue: Issue
    let statuses: [Status]
    let members: [OrgMember]
    var onSave: (String, String, String?) async -> Bool

    @Environment(\.dismiss) private var dismiss
    @State private var title: String
    @State private var selectedStatusId: String
    @State private var selectedAssigneeId: String
    @State private var isSubmitting = false
    @State private var errorMessage: String?

    init(
        issue: Issue,
        statuses: [Status],
        members: [OrgMember],
        onSave: @escaping (String, String, String?) async -> Bool
    ) {
        self.issue = issue
        self.statuses = statuses
        self.members = members
        self.onSave = onSave
        _title = State(initialValue: issue.title)
        _selectedStatusId = State(initialValue: issue.statusId)
        _selectedAssigneeId = State(initialValue: issue.assigneeId ?? "")
    }

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Title", text: $title)
                }

                Section {
                    Picker("Status", selection: $selectedStatusId) {
                        ForEach(statuses) { status in
                            Text(status.name).tag(status.id)
                        }
                    }
                    Picker("Assignee", selection: $selectedAssigneeId) {
                        Text("Unassigned").tag("")
                        ForEach(members) { member in
                            Text(member.displayName).tag(member.id)
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
            .navigationTitle("Edit issue")
            .navigationBarTitleDisplayMode(.inline)
            .toolbar {
                ToolbarItem(placement: .cancellationAction) {
                    Button("Cancel") { dismiss() }
                }
                ToolbarItem(placement: .confirmationAction) {
                    Button(isSubmitting ? "Saving…" : "Save") {
                        Task { await save() }
                    }
                    .disabled(
                        isSubmitting
                            || title.trimmingCharacters(in: .whitespaces).isEmpty
                            || selectedStatusId.isEmpty
                    )
                }
            }
        }
    }

    private func save() async {
        isSubmitting = true
        errorMessage = nil
        defer { isSubmitting = false }

        let assigneeId = selectedAssigneeId.isEmpty ? nil : selectedAssigneeId
        let success = await onSave(
            title.trimmingCharacters(in: .whitespaces),
            selectedStatusId,
            assigneeId
        )
        if success {
            dismiss()
        } else {
            errorMessage = "Failed to update issue"
        }
    }
}
