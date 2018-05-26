import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute, Params } from '@angular/router';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';

import { Collection, Shard, BackendService } from '../backend.service';
import { UserComponent } from '../user/user.component';

@Component({
  selector: 'rana-collection-settings',
  templateUrl: './settings.component.html',
})
export class CollectionSettingsComponent implements OnInit {
  form: FormGroup;
  collection: Collection;

  shards: Shard[];
  allSize: number;

  constructor(
    private backend: BackendService,
    private router: Router,
    private route: ActivatedRoute,
    private fb: FormBuilder,
    private user: UserComponent,
  ) { }

  ngOnInit() {
    this.form = this.fb.group({
      id: [null],
      name: [null, [Validators.required]],
    });
    this.route.parent.params.forEach((params: Params) => {
      const collectionId = params['collectionId'];
      this.getCollection(collectionId);
      this.getCollectionShards(collectionId);
    });
  }

  getCollection(collectionId: string) {
    this.backend
      .getCollection(this.user.user, collectionId)
      .subscribe(collection => {
        this.form.setValue(collection);
        this.collection = collection;
      });
  }

  getCollectionShards(collectionId: string) {
   this.backend
      .getCollectionShards(this.user.user, collectionId)
      .subscribe(shards => {
        this.shards = shards;
        this.allSize = shards.reduce((a, b) => a + b.size, 0);
      });
  }

  save() {
    this.backend.saveCollection(this.user.user, this.form.value)
      .subscribe(_ => this.router.navigate(['..'], {relativeTo: this.route}));
  }

  delete() {
    this.backend.deleteCollection(this.user.user, this.form.value.id)
      .subscribe(_ => this.router.navigate(['../..'], {relativeTo: this.route}));
  }

  deleteShard(shard: Shard) {
    this.backend
      .deleteCollectionShard(this.user.user, this.collection.id, shard.id)
      .subscribe(_ => {
        this.getCollectionShards(this.collection.id);
      })
  }

}
