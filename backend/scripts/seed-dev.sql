-- Development seed: admin user, workspace, demo project and sample issues.
-- Run via: make seed-dev
-- Login: admin@plexus.local / admin  (or type "admin" on the login form)

DO $$
DECLARE
    v_user_id       UUID := '00000000-0000-0000-0000-000000000001';
    v_org_id        UUID := '00000000-0000-0000-0000-000000000010';
    v_project_id    UUID := '00000000-0000-0000-0000-000000000020';
    v_sprint_id     UUID := '00000000-0000-0000-0000-000000000030';
    v_todo_status   UUID;
    v_progress      UUID;
    v_review        UUID;
    v_done          UUID;
    v_task_type     UUID;
    v_bug_type      UUID;
BEGIN
    INSERT INTO users (id, email, password_hash, display_name, role)
    VALUES (
        v_user_id,
        'admin@plexus.local',
        '$2a$10$a4Mp8EmXNunz4DRO//NNPu1RsHrr6w2/6gY4ZERxee6zexNW/nLbC',
        'Admin',
        'admin'
    )
    ON CONFLICT (email) DO NOTHING;

    INSERT INTO organizations (id, slug, name, plan)
    VALUES (v_org_id, 'plexus', 'Plexus Dev', 'free')
    ON CONFLICT (slug) DO NOTHING;

    INSERT INTO org_members (org_id, user_id, role)
    VALUES (v_org_id, v_user_id, 'owner')
    ON CONFLICT (org_id, user_id) DO NOTHING;

    INSERT INTO projects (id, org_id, key, name, description, lead_id)
    VALUES (v_project_id, v_org_id, 'DEMO', 'Demo Project', 'Sample project for development', v_user_id)
    ON CONFLICT (org_id, key) DO NOTHING;

    IF NOT EXISTS (SELECT 1 FROM statuses WHERE project_id = v_project_id) THEN
        PERFORM seed_project_defaults(v_project_id);
    END IF;

    SELECT id INTO v_todo_status FROM statuses WHERE project_id = v_project_id AND name = 'To Do';
    SELECT id INTO v_progress   FROM statuses WHERE project_id = v_project_id AND name = 'In Progress';
    SELECT id INTO v_review     FROM statuses WHERE project_id = v_project_id AND name = 'In Review';
    SELECT id INTO v_done       FROM statuses WHERE project_id = v_project_id AND name = 'Done';
    SELECT id INTO v_task_type  FROM issue_types WHERE project_id = v_project_id AND name = 'Task';
    SELECT id INTO v_bug_type   FROM issue_types WHERE project_id = v_project_id AND name = 'Bug';

    INSERT INTO sprints (id, project_id, name, goal, state, start_date, end_date)
    VALUES (
        v_sprint_id,
        v_project_id,
        'Sprint 1',
        'Initial development sprint',
        'active',
        NOW() - INTERVAL '3 days',
        NOW() + INTERVAL '11 days'
    )
    ON CONFLICT (id) DO NOTHING;

    IF NOT EXISTS (SELECT 1 FROM issues WHERE project_id = v_project_id) THEN
        INSERT INTO issues (project_id, number, type_id, status_id, title, description, priority, assignee_id, reporter_id, sprint_id, story_points, position) VALUES
            (v_project_id, 1, v_task_type, v_todo_status, 'Set up project board', 'Configure columns and workflow for the team.', 'medium', v_user_id, v_user_id, v_sprint_id, 3, 1),
            (v_project_id, 2, v_bug_type,  v_progress,   'Fix login redirect', 'Users are not redirected after sign-in.', 'high', v_user_id, v_user_id, v_sprint_id, 2, 2),
            (v_project_id, 3, v_task_type, v_review,     'Design issue detail page', 'Create layout for issue view with comments.', 'medium', NULL, v_user_id, v_sprint_id, 5, 3),
            (v_project_id, 4, v_task_type, v_done,       'Initialize repository', 'Monorepo with backend and clients.', 'low', v_user_id, v_user_id, v_sprint_id, 1, 4),
            (v_project_id, 5, v_task_type, v_todo_status, 'Add sprint planning', 'Backlog grooming and sprint creation.', 'medium', NULL, v_user_id, NULL, 8, 5);
    END IF;
END $$;
