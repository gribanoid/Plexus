import SwiftUI

struct RegisterView: View {
    @EnvironmentObject var authStore: AuthStore
    @State private var displayName = ""
    @State private var email = ""
    @State private var password = ""
    @Environment(\.dismiss) private var dismiss

    var body: some View {
        VStack(spacing: 24) {
            Spacer()

            VStack(spacing: 8) {
                Text("Create account")
                    .font(.largeTitle.bold())
                Text("Start managing your projects")
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
            }

            VStack(spacing: 12) {
                TextField("Full name", text: $displayName)
                    .textContentType(.name)
                    .textFieldStyle(.roundedBorder)

                TextField("Work email", text: $email)
                    .textContentType(.emailAddress)
                    .keyboardType(.emailAddress)
                    .autocapitalization(.none)
                    .textFieldStyle(.roundedBorder)

                SecureField("Password (min 8 chars)", text: $password)
                    .textContentType(.newPassword)
                    .textFieldStyle(.roundedBorder)

                if let error = authStore.errorMessage {
                    Text(error)
                        .font(.caption)
                        .foregroundStyle(.red)
                        .frame(maxWidth: .infinity, alignment: .leading)
                }

                Button {
                    Task {
                        await authStore.register(
                            email: email, password: password, displayName: displayName
                        )
                    }
                } label: {
                    HStack {
                        if authStore.isLoading { ProgressView().tint(.white) }
                        Text(authStore.isLoading ? "Creating…" : "Create account")
                            .font(.headline)
                    }
                    .frame(maxWidth: .infinity)
                    .padding(.vertical, 12)
                    .background(.primary)
                    .foregroundStyle(.background)
                    .clipShape(RoundedRectangle(cornerRadius: 10))
                }
                .disabled(authStore.isLoading || displayName.isEmpty || email.isEmpty || password.count < 8)
            }

            Button("Already have an account? Sign in") { dismiss() }
                .font(.subheadline)
                .foregroundStyle(.secondary)

            Spacer()
        }
        .padding(.horizontal, 24)
        .navigationBarBackButtonHidden(true)
    }
}
