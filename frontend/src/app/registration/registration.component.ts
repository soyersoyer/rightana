import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';

import { AuthService, BackendService } from '../backend.service';

@Component({
  selector: 'rana-registration',
  templateUrl: './registration.component.html',
})
export class RegistrationComponent implements OnInit {
  form: FormGroup;

  constructor(
    private fb: FormBuilder,
    private router: Router,
    private auth: AuthService,
    private backend: BackendService,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      name: [null, [Validators.required, Validators.pattern("^[a-z0-9.]+$")]],
      email: [null, [Validators.required, Validators.email]],
      password: [null, [Validators.required]]
    });
  }

  registrate() {
    this.backend
      .createUser(this.form.value)
      .subscribe(() => this.login());
  }

  login() {
    this.backend
      .createAuthToken(this.form.value)
      .subscribe(token => {
        this.auth.set(token.id, token.user_info.name, token.user_info.is_admin);
        this.router.navigateByUrl(token.user_info.name);
      });
  }

}
