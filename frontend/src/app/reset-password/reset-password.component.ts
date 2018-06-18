import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup } from '@angular/forms';
import { ActivatedRoute, Params, Router } from '@angular/router';

import { UserComponent } from '../user/user.component';
import { BackendService, AuthService } from "../backend.service"
import { RValidators } from '../forms/rvalidators';
import { ToastyService } from '../toasty/toasty.module';

@Component({
  selector: 'rana-reset-password',
  templateUrl: './reset-password.component.html',
})
export class ResetPasswordComponent implements OnInit {
  form: FormGroup;
  resetKey: string;

  constructor(
    private user: UserComponent,
    private fb: FormBuilder,
    private backend: BackendService,
    private auth: AuthService,
    private toasty: ToastyService,
    private route: ActivatedRoute,
    private router: Router,
  ) { }

  ngOnInit() {
    this.route.queryParams.forEach((params: Params) => {
      this.resetKey = params['reset_key'];
    });
    this.form = this.fb.group({
      password: [null, RValidators.password],
    });
  }

  resetPassword() {
    this.backend
      .resetPassword(this.user.user, this.resetKey, this.form.value.password)
      .subscribe(_ => {
        this.login();
        this.toasty.success("Password change success");
      });
  }

  login() {
    this.backend
      .createAuthToken(this.user.user, this.form.value.password)
      .subscribe(token => {
        this.auth.set(token.id, token.user_info.name, token.user_info.is_admin);
        this.router.navigateByUrl('/'+token.user_info.name);
      });
  }

}
