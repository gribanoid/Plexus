import SwiftUI

@MainActor
final class AuthStore: ObservableObject {
    @Published var currentUser: User?
    @Published var isLoading = false
    @Published var errorMessage: String?

    private let api = APIClient.shared
    private let keychain = KeychainStore.shared

    var isAuthenticated: Bool {
        keychain.accessToken != nil
    }

    func login(email: String, password: String) async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            struct LoginBody: Encodable { let email: String; let password: String }
            let pair: TokenPair = try await api.request(
                "auth/login",
                method: "POST",
                body: LoginBody(email: email, password: password)
            )
            keychain.accessToken = pair.accessToken
            keychain.refreshToken = pair.refreshToken
            await fetchMe()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func register(email: String, password: String, displayName: String) async {
        isLoading = true
        errorMessage = nil
        defer { isLoading = false }

        do {
            struct RegisterBody: Encodable {
                let email: String
                let password: String
                let displayName: String
            }
            let pair: TokenPair = try await api.request(
                "auth/register",
                method: "POST",
                body: RegisterBody(email: email, password: password, displayName: displayName)
            )
            keychain.accessToken = pair.accessToken
            keychain.refreshToken = pair.refreshToken
            await fetchMe()
        } catch {
            errorMessage = error.localizedDescription
        }
    }

    func logout() {
        keychain.accessToken = nil
        keychain.refreshToken = nil
        currentUser = nil
    }

    func fetchMe() async {
        do {
            currentUser = try await api.request("me")
        } catch APIError.unauthorized {
            logout()
        } catch {
            // Non-fatal: user info unavailable
        }
    }
}
