import { Component, OnInit } from '@angular/core';

import { UserInfo, BackendService } from '../backend.service';

@Component({
  selector: 'k20a-admin-users',
  templateUrl: './admin-users.component.html',
})
export class AdminUsersComponent implements OnInit {
  users: UserInfo[];

  constructor(
    private backend: BackendService,
  ) { }

  ngOnInit() {
    this.getUsers();
  }

  getUsers() {
    this.backend.getUsers().subscribe(users => this.users = users)
  }
}
