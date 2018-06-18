import { Component, OnInit } from '@angular/core';

import { BackendService, AuthService, UserInfo } from "../../backend.service"
import { ToastyService } from '../../toasty/toasty.module';

@Component({
  selector: 'rana-profile',
  templateUrl: './profile.component.html',
})
export class ProfileComponent implements OnInit {
  user: UserInfo;

  constructor(
    private backend: BackendService,
    private auth: AuthService,
    private toasty: ToastyService,
  ) { }

  ngOnInit() {
    this.backend.getUserInfo(this.auth.user)
    .subscribe(user => {
      this.user = user;
    });
  }

  sendVerifyEmail() {
    this.backend.sendVerifyEmail(this.auth.user)
    .subscribe(_ => {
      this.toasty.success('Verify Email sent!')
    });
  }
}
