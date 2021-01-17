
{{template "header"}}

<ul class="nav nav-tabs mt-3">
  <li class="nav-item">
    <a class="nav-link" id="settings-tab" href="/admin/settings" role="tab" aria-controls="settings" aria-selected="false">General Settings</a>
  </li>
  <li class="nav-item">
      <a class="nav-link" id="apps-tab" href="/admin/apps" role="tab" aria-controls="apps" aria-selected="false">Apps</a>
  </li>
  <li class="nav-item">
      <a class="nav-link active" id="users-tab" href="/admin/users" role="tab" aria-controls="users" aria-selected="true">Users</a>
  </li>
  <li class="nav-item">
      <a class="nav-link" id="groups-tab" href="/admin/groups" role="tab" aria-controls="groups" aria-selected="false">Groups</a>
  </li>
</ul>
<div class="tab-content p-3" id="adminTabContent">
    <div class="tab-pane active" id="users" role="tabpanel" aria-labelledby="users-tab">...</div>
</div>

{{template "footer"}}
