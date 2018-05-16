import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { Router } from '@angular/router';

import { AuthService, BackendService, K20Config } from '../backend.service';

@Component({
  selector: 'k20a-login',
  templateUrl: './login.component.html',
})
export class LoginComponent implements OnInit {
  form: FormGroup;
  config: K20Config;

  constructor(
    private fb: FormBuilder,
    private backend: BackendService,
    private auth: AuthService,
    private router: Router,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      email: [null, [Validators.required, Validators.email]],
      password: [null, [Validators.required]]
    });
    this.getConfig();
  }

  getConfig() {
    this.backend.getConfig()
      .subscribe(config => this.config = config);
  }

  login() {
    this.backend
      .createAuthToken(this.form.value)
      .subscribe(token => {
        this.auth.set(token.id, token.user_info.name, token.user_info.is_admin);
        this.router.navigateByUrl('/collections');
      });
  }
}
