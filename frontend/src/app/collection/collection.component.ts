import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute, Params } from '@angular/router';
import { Observable } from 'rxjs/Observable';
import { tap } from 'rxjs/operators';

import { UserComponent } from '../user/user.component';
import { BackendService, CollectionSummary, AuthService } from '../backend.service';

@Component({
  selector: 'rana-collection',
  templateUrl: './collection.component.html',
})
export class CollectionComponent implements OnInit {
  collections: CollectionSummary[];

  constructor(
    private backend: BackendService,
    private route: ActivatedRoute,
    private router: Router,
    private userComp: UserComponent,
    private auth: AuthService,
  ) { }

  ngOnInit() {
    this.route.params.forEach((params: Params) => {
      this.getCollections(this.user);
    })
  }

  get user(): string {
    return this.userComp.user;
  }

  get selfPage(): boolean {
    return this.user === this.auth.user
  }

  getCollections(user: string) {
    this.backend.getCollectionSummaries(user).subscribe(collections => this.collections = collections);
  }


}
