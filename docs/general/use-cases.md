# Use cases by actor

## Organization owner

| Goal | Main flow | Result |
| --- | --- | --- |
| Create workspace | Register/login -> create organization -> become owner | A tenant workspace exists and owner membership is created. |
| Invite team member | Open organization members -> invite/add user -> assign role | User can access permitted organization resources. |
| Maintain ownership | Transfer ownership deliberately before leaving | Organization always has an owner. |

## Admin

| Goal | Main flow | Result |
| --- | --- | --- |
| Manage members | View organization members -> add/remove/change allowed role | Membership changes are audited. |
| Review organization work | Select project -> view authorized project data | Admin obtains visibility without bypassing tenant isolation. |

## Project manager

| Goal | Main flow | Result |
| --- | --- | --- |
| Create project | Provide name/description -> choose members | Project and its default group-chat channel are created. |
| Plan work | Create task -> assign priority, assignee, due date, status | Team receives one accountable work item. |
| Track delivery | Open board -> filter tasks -> review activity | Manager sees current work state and blockers. |
| Collaborate | Comment or send project chat message | Project members receive a persisted real-time update. |

## Member

| Goal | Main flow | Result |
| --- | --- | --- |
| View assigned work | Open project or personal task list | Sees only tasks in accessible projects. |
| Progress task | Open task -> move permitted status -> submit update | Change is validated, logged, and broadcast. |
| Discuss work | Add task comment or use project chat | Conversation remains attached to the relevant project/task. |

## System / worker

| Goal | Main flow | Result |
| --- | --- | --- |
| Deliver reliable event | Read committed outbox event -> publish -> mark handled | Real-time clients can synchronize after durable data is saved. |
| Run deferred work | Receive queued job -> execute idempotently -> retry on transient error | User-facing HTTP requests remain fast. |

## Authorization decision for every use case

1. Identify authenticated user.
2. Identify requested organization from trusted route/resource context.
3. Confirm active organization membership.
4. Confirm project membership or allowed organization role.
5. Confirm action permission and ownership rule.
6. Execute and audit the approved action.
