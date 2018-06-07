import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { ToastyService } from 'ng2-toasty';

import { BackendService, AuthService } from "../../backend.service"
import { RValidators } from '../../forms/rvalidators';

@Component({
  selector: 'rana-change-password',
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
      currentPassword: [null, RValidators.password],
      password: [null, RValidators.password],
    });
  }

  changePassword() {
    const v = this.form.value
    this.backend
      .updateUserPassword(this.auth.user, v.currentPassword, v.password)
      .subscribe(_ => {
        this.form.reset();
        this.toasty.success("Password change success");
      });
  }
}
