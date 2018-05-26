import { Component, OnInit } from '@angular/core';
import { Router, ActivatedRoute, Params } from '@angular/router';
import { Observable } from 'rxjs/Observable';
import { tap } from 'rxjs/operators';

import { UserComponent } from '../user/user.component';
import { BackendService, CollectionSummary } from '../backend.service';

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
  ) { }

  ngOnInit() {
    this.route.params.forEach((params: Params) => {
      this.getCollections(this.user)
      .subscribe(_ => {
        if (this.route.snapshot.children.length === 0) {
          this.navigateToCreateIfEmpty();
        };
      });
    })
  }

  get user(): string {
    return this.userComp.user;
  }

  getCollections(user: string): Observable<CollectionSummary[]> {
    return this.backend.getCollectionSummaries(user)
      .do(collections => this.collections = collections);
  }


  navigateToCreateIfEmpty() {
    if (this.collections && this.collections.length === 0) {
     this.router.navigate(['create'], {relativeTo: this.route});
    }
  }
}
