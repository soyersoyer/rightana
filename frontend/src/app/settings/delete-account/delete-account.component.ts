import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';

import { ToastyService } from 'ng2-toasty';

import { BackendService, AuthService } from "../../backend.service"

@Component({
  selector: 'k20a-delete-account',
  templateUrl: './delete-account.component.html',
})
export class DeleteAccountComponent implements OnInit {
  form: FormGroup;

  constructor(
    private fb: FormBuilder,
    private backend: BackendService,
    private auth: AuthService,
    private toasty: ToastyService,
    private router: Router,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      password: [null, [Validators.required]],
    });
  }

  deleteAccount() {
    this.backend
      .deleteUser(this.auth.email, this.form.value.password)
      .subscribe(_ => {
        this.toasty.success('Account delete success');
        this.auth.unset();
        this.router.navigate(['/']);
      });
  }
}
