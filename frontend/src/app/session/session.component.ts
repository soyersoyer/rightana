import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Params, Router } from '@angular/router';

import { BackendService, Session, Pageview } from '../backend.service';

import { CollectionDashboardComponent, Setup } from '../collection-dashboard/dashboard.component';

class SessionD extends Session {
  pageviews?: Pageview[];
  showDetails?: boolean;
}

@Component({
  selector: 'k20a-session',
  templateUrl: './session.component.html',
})
export class SessionComponent implements OnInit, OnDestroy {
  sessions: SessionD[];

  setup = new Setup();

  subscription: any;

  actual = 50;

  constructor(
    private backend: BackendService,
    private dashboard: CollectionDashboardComponent,
  ) { }

  ngOnInit() {
    this.getSessions(this.dashboard.setup);
    this.subscription = this.dashboard.setup.events.subscribe(setup => {
      this.getSessions(setup);
    });
  }

  ngOnDestroy() {
    this.subscription.unsubscribe();
  }

  getSessions(setup: Setup) {
    this.backend
      .getSessions(setup.collectionId, setup.from, setup.to, setup.filter)
      .subscribe(sessions => {
        this.sessions = sessions;
        this.actual = 50;
        this.setup.set(setup);
      });
  }

  toggleDetails(session: SessionD) {
    if (!session.showDetails) {
      session.showDetails = true;
      this.getPageviews(session)
    } else {
      session.showDetails = false;
    }
  }

  getPageviews(session: SessionD) {
    this.backend
      .getPageviews(this.setup.collectionId, session.key)
      .subscribe(pageviews => session.pageviews = pageviews);
  }

  loadMore() {
    this.actual += 50;
  }
}

