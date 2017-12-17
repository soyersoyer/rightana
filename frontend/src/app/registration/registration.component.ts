import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';

import { AuthService, BackendService } from '../backend.service';

@Component({
  selector: 'k20a-registration',
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
        this.auth.set(token.id, this.form.value.email);
        this.router.navigateByUrl('/collections');
      });
  }

}
