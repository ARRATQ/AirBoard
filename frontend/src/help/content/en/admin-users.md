# Users

The **Administration → Users** page creates and manages accounts. People who sign in with Microsoft get their account created on their first sign-in.

## Creating an account

1. Open [Users](/admin/users) and click **New user**.
2. **Account** tab: **Username** (it cannot change later), **Email address**, first name, last name and **Password**. Keep **Active account** ticked.
3. **Role & groups** tab:
   - **Role**: **User** (access to their groups' resources) or **Administrator** (access to everything);
   - **Groups**: they decide what the person sees;
   - **Assigned roles**: the roles created on the Roles page (see the "Roles and permissions" guide).
4. Click **Create user**.

## Limiting a role to some groups

Under **Assigned roles**, each role offers **Groups covered by the role**. With no choice, the role applies to the groups the person belongs to (except the default group). Pick groups to narrow it.

## Suspending or deleting

- **Deactivate**: the person can no longer sign in; **Activate** restores access.
- **Delete** moves the account to the **Trash**: **Restore user** brings it back.
- From the trash, **Delete permanently** erases the account and its data. This cannot be undone.

> **Good to know**: groups that come from the directory are **locked** in the user form. They follow the person's unit: change their assignment in the directory to change them.
