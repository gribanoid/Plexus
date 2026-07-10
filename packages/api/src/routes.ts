export const routes = {
  orgs: () => '/orgs',
  orgNew: () => '/orgs/new',
  org: (orgSlug: string) => `/orgs/${orgSlug}`,
  projectBoard: (orgSlug: string, projectKey: string) =>
    `/orgs/${orgSlug}/${projectKey}/board`,
  projectBacklog: (orgSlug: string, projectKey: string) =>
    `/orgs/${orgSlug}/${projectKey}/backlog`,
  projectSettings: (orgSlug: string, projectKey: string) =>
    `/orgs/${orgSlug}/${projectKey}/settings`,
  issue: (orgSlug: string, projectKey: string, issueNumber: number | string) =>
    `/orgs/${orgSlug}/${projectKey}/issues/${issueNumber}`,
}
