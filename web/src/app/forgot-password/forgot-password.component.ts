import { Component, OnInit } from '@angular/core';
import { FormGroup, FormBuilder, Validators } from '@angular/forms';

import { RValidators } from '../forms/rvalidators';
import { BackendService } from '../backend.service';
import { ToastyService } from '../toasty/toasty.module';

@Component({
  selector: 'rana-forgot-password',
  templateUrl: './forgot-password.component.html',
})
export class ForgotPasswordComponent implements OnInit {
  form: FormGroup;

  constructor(
    private fb: FormBuilder,
    private backend: BackendService,
    private toasty: ToastyService,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      email: [null, RValidators.email],
    });
  }

  send() {
    this.backend
      .sendResetPassword(this.form.value.email)
      .subscribe(_ => {
        this.toasty.success('Reset password email sent!')
        this.form.reset();
      });
  }

}
