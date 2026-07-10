import SwiftUI

struct LoginView: View {
    @EnvironmentObject var authStore: AuthStore
    @State private var email = ""
    @State private var password = ""
    @State private var showRegister = false

    var body: some View {
        NavigationStack {
            VStack(spacing: 24) {
                Spacer()

                // Logo
                VStack(spacing: 8) {
                    RoundedRectangle(cornerRadius: 16)
                        .fill(.primary)
                        .frame(width: 64, height: 64)
                        .overlay(
                            Text("P")
                                .font(.system(size: 32, weight: .bold))
                                .foregroundColor(.white)
                        )
                    Text("Plexus")
                        .font(.largeTitle.bold())
                    Text("Sign in to your workspace")
                        .font(.subheadline)
                        .foregroundStyle(.secondary)
                }

                // Form
                VStack(spacing: 12) {
                    TextField("Work email", text: $email)
                        .textContentType(.emailAddress)
                        .keyboardType(.emailAddress)
                        .autocapitalization(.none)
                        .textFieldStyle(.roundedBorder)

                    SecureField("Password", text: $password)
                        .textContentType(.password)
                        .textFieldStyle(.roundedBorder)

                    if let error = authStore.errorMessage {
                        Text(error)
                            .font(.caption)
                            .foregroundStyle(.red)
                            .frame(maxWidth: .infinity, alignment: .leading)
                    }

                    Button {
                        Task { await authStore.login(email: email, password: password) }
                    } label: {
                        HStack {
                            if authStore.isLoading {
                                ProgressView().tint(.white)
                            }
                            Text(authStore.isLoading ? "Signing in…" : "Sign in")
                                .font(.headline)
                        }
                        .frame(maxWidth: .infinity)
                        .padding(.vertical, 12)
                        .background(.primary)
                        .foregroundStyle(.background)
                        .clipShape(RoundedRectangle(cornerRadius: 10))
                    }
                    .disabled(authStore.isLoading || email.isEmpty || password.isEmpty)
                }

                Button("Don't have an account? Create one") {
                    showRegister = true
                }
                .font(.subheadline)
                .foregroundStyle(.secondary)

                Spacer()
            }
            .padding(.horizontal, 24)
            .navigationDestination(isPresented: $showRegister) {
                RegisterView()
            }
        }
    }
}
