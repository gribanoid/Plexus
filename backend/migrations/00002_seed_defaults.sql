-- +goose Up
-- Function to seed default issue types and statuses for a new project
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION seed_project_defaults(p_project_id UUID)
RETURNS void AS $$
BEGIN
    -- Default issue types
    INSERT INTO issue_types (id, project_id, name, color) VALUES
        (uuid_generate_v4(), p_project_id, 'Task',  '#3B82F6'),
        (uuid_generate_v4(), p_project_id, 'Bug',   '#EF4444'),
        (uuid_generate_v4(), p_project_id, 'Story', '#8B5CF6'),
        (uuid_generate_v4(), p_project_id, 'Epic',  '#F59E0B');

    -- Default statuses
    INSERT INTO statuses (id, project_id, name, color, category, position) VALUES
        (uuid_generate_v4(), p_project_id, 'To Do',       '#6B7280', 'todo',        0),
        (uuid_generate_v4(), p_project_id, 'In Progress', '#3B82F6', 'in_progress', 1),
        (uuid_generate_v4(), p_project_id, 'In Review',   '#F59E0B', 'in_progress', 2),
        (uuid_generate_v4(), p_project_id, 'Done',        '#10B981', 'done',        3);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
DROP FUNCTION IF EXISTS seed_project_defaults(UUID);
