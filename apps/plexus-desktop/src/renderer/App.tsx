import { HashRouter, Routes, Route, Navigate } from 'react-router-dom'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { OrgsPage } from './pages/OrgsPage'
import { NewOrgPage } from './pages/NewOrgPage'
import { OrgProjectsPage } from './pages/OrgProjectsPage'
import { BoardPage } from './pages/BoardPage'
import { BacklogPage } from './pages/BacklogPage'
import { IssuePage } from './pages/IssuePage'
import { SettingsPage } from './pages/SettingsPage'
import { AppShell } from './components/AppShell'
import { RequireAuth } from './components/RequireAuth'

export function App() {
  return (
    <HashRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route
          path="/"
          element={
            <RequireAuth>
              <AppShell />
            </RequireAuth>
          }
        >
          <Route index element={<Navigate to="/orgs" replace />} />
          <Route path="orgs" element={<OrgsPage />} />
          <Route path="orgs/new" element={<NewOrgPage />} />
          <Route path="orgs/:orgSlug" element={<OrgProjectsPage />} />
          <Route path="orgs/:orgSlug/:projectKey/board" element={<BoardPage />} />
          <Route path="orgs/:orgSlug/:projectKey/backlog" element={<BacklogPage />} />
          <Route path="orgs/:orgSlug/:projectKey/issues/:issueNumber" element={<IssuePage />} />
          <Route path="orgs/:orgSlug/:projectKey/settings" element={<SettingsPage />} />
        </Route>
        <Route path="*" element={<Navigate to="/orgs" replace />} />
      </Routes>
    </HashRouter>
  )
}
