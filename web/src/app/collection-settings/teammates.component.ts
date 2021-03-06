import { Component, OnInit, Input } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { Collection, BackendService, Teammate } from '../backend.service';
import { UserComponent } from '../user/user.component';
import { RValidators } from '../forms/rvalidators';

@Component({
  selector: 'rana-collection-teammates',
  templateUrl: './teammates.component.html',
})
export class TeammatesComponent implements OnInit {
  form: FormGroup;
  @Input() collection: Collection;
  teammates: Teammate[];

  constructor(
    private backend: BackendService,
    private fb: FormBuilder,
    private user: UserComponent,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      email: [null, RValidators.email],
    });
    this.getTeammates();
  }

  getTeammates() {
    this.backend
      .getTeammates(this.user.user, this.collection.name)
      .subscribe(teammates => this.teammates = teammates);
  }

  add() {
    this.backend
      .addTeammate(this.user.user, this.collection.name, this.form.value.email)
      .subscribe(_ => {
        this.getTeammates();
        this.form.reset();
      });
  }

  remove(email: string) {
    this.backend
      .removeTeammate(this.user.user, this.collection.name, email)
      .subscribe(_ => this.getTeammates());
  }
}
