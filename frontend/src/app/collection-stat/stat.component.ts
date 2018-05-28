import { Component, OnInit, OnDestroy } from '@angular/core';

import { BackendService, CollectionSumData } from '../backend.service';

import { CollectionDashboardComponent, Setup } from '../collection-dashboard/dashboard.component';

@Component({
  selector: 'rana-collection-stat',
  templateUrl: './stat.component.html',
})
export class CollectionStatComponent implements OnInit, OnDestroy {
  sums: CollectionSumData;

  dashboardSetup: Setup;
  setup = new Setup();

  subscription: any;

  constructor(
    private backend: BackendService,
    private dashboard: CollectionDashboardComponent,
  ) { }

  ngOnInit() {
    this.dashboardSetup = this.dashboard.setup;
    this.getSums(this.dashboard.setup);
    this.subscription = this.dashboard.setup.events.subscribe(setup => {
      this.getSums(setup);
    });
  }

  getSums(setup: Setup) {
    this.backend
      .getCollectionStatData(this.dashboard.user, setup.collectionName, setup.from, setup.to, setup.filter)
      .subscribe(sums => {
        this.sums = sums;
        this.setup.set(setup);
      });
  }

  ngOnDestroy() {
    this.subscription.unsubscribe();
  }

}
