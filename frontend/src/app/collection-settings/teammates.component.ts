import { Component, OnInit, Input } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { Collection, BackendService, Teammate } from '../backend.service';

@Component({
  selector: 'k20a-collection-teammates',
  templateUrl: './teammates.component.html',
})
export class TeammatesComponent implements OnInit {
  form: FormGroup;
  @Input() collection: Collection;
  teammates: Teammate[];

  constructor(
    private backend: BackendService,
    private fb: FormBuilder,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      email: [null, [Validators.required]],
    });
    this.getTeammates();
  }

  getTeammates() {
    this.backend
      .getTeammates(this.collection.id)
      .subscribe(teammates => this.teammates = teammates);
  }

  add() {
    this.backend
      .addTeammate(this.collection.id, this.form.value.email)
      .subscribe(_ => {
        this.getTeammates();
        this.form.reset();
      });
  }

  remove(email: string) {
    this.backend
      .removeTeammate(this.collection.id, email)
      .subscribe(_ => this.getTeammates());
  }
}
