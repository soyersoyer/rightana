import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { ToastyService } from 'ng2-toasty';

import { BackendService, AuthService } from "../../backend.service"

@Component({
  selector: 'k20a-change-password',
  templateUrl: './change-password.component.html',
})
export class ChangePasswordComponent implements OnInit {
  form: FormGroup;

  constructor(
  	private fb: FormBuilder,
    private backend: BackendService,
    private auth: AuthService,
    private toasty: ToastyService,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      currentPassword: [null, [Validators.required]],
      password: [null, [Validators.required]],
      password2: [null, [Validators.required]],
    });
  }

  changePassword() {
    const v = this.form.value
    this.backend
      .updateUserPassword(this.auth.email, v.currentPassword, v.password)
      .subscribe(_ => {
        this.form.reset();
        this.toasty.success("Password change success");
      });
  }
}