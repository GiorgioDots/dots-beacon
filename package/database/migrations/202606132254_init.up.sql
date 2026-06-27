CREATE TABLE
    users (
        id UUID NOT NULL DEFAULT uuidv7 (),
        external_id TEXT NOT NULL,
        PRIMARY KEY (id)
    );

CREATE TABLE
    sites (
        id UUID NOT NULL DEFAULT uuidv7 (),
        name TEXT NOT NULL,
        is_on BOOLEAN NOT NULL DEFAULT FALSE,
        PRIMARY KEY (id)
    );

CREATE TABLE
    sites_users (
        id UUID NOT NULL DEFAULT uuidv7 (),
        user_id UUID NOT NULL REFERENCES users (id),
        site_id UUID NOT NULL REFERENCES sites (id),
        PRIMARY KEY (id)
    );

CREATE TABLE 
    sites_permission_groups (
        id UUID NOT NULL DEFAULT uuidv7 (),
        site_id UUID NOT NULL REFERENCES sites (id),
        name TEXT NOT NULL,
        permission TEXT NOT NULL, -- fake reference to permission table. the values will be like user.* or * or super_admin.create_site
        PRIMARY KEY (id)
    );

CREATE TABLE 
    sites_users_permission_groups (
        id UUID NOT NULL DEFAULT uuidv7 (),
        site_user_id UUID NOT NULL REFERENCES sites_users (id),
        site_permission_group_id UUID NOT NULL REFERENCES sites_permission_groups (id),
        PRIMARY KEY (id) 
    );

-- this table will change only if there're new features to be under permission 
CREATE TABLE 
    permissions (
        key TEXT NOT NULL, -- like super_admin.create_site or user.view_events
        description TEXT,
        is_global BOOLEAN,
        PRIMARY KEY (key)
    );

INSERT INTO permissions (key, description, is_global) VALUES 
    ('super_admin.sites.create', 'Create new sites', TRUE),
    ('super_admin.sites.edit', 'Edit sites', TRUE),
    ('super_admin.sites.delete', 'Permanently delete a site', TRUE),
    ('super_admin.users.create', 'Create a user in a site', TRUE),
    ('super_admin.users.edit_permissions', 'Edit the users permissions in a site', TRUE),
    ('super_admin.users.remove', 'Remove a user from a site', TRUE)

