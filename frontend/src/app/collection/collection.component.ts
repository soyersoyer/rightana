import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

import { UserComponent } from '../user/user.component';
import { BackendService, CollectionSummary, AuthService } from '../backend.service';
import { getDateStrFromUnixTime } from '../utils/date';

@Component({
  selector: 'rana-collection',
  templateUrl: './collection.component.html',
})
export class CollectionComponent implements OnInit {
  collections: CollectionSummary[];

  timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  data: {
    labels: string[],
    datasets: {
      label: string,
      borderColor: string,
      fill: boolean,
      data: {x: string, y: number}[],
    }[],
  }[];
  options = {
    tooltips: {mode: 'index', intersect: false},
    hover: {mode: 'nearest', intersect: true},
    legend: {display: false},
    layout: {padding: 5},
    elements: {point: {pointStyle: 'star'}},
    scales: {
      xAxes: [{
        display: false,
        stacked: true
      }],
      yAxes: [
        {
          id: 'session-axis',
          display: false,
          ticks: {
            beginAtZero: true,
          },
        },
        {
          id: 'page-view-axis',
          display: false,
          ticks: {
            beginAtZero: true,
          },
        },
      ]
    }
  };

  constructor(
    private backend: BackendService,
    private route: ActivatedRoute,
    private userComp: UserComponent,
    private auth: AuthService,
  ) { }

  ngOnInit() {
    this.route.params.forEach(_ => {
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
    this.backend.getCollectionSummaries(user, this.timezone)
      .subscribe(collections => {
        this.collections = collections;
        this.setData(collections);
      });
  }

  setData(collections: CollectionSummary[]) {
    this.data = collections.map(c => ({
      labels: c.session_sums.map(s => getDateStrFromUnixTime(s.bucket, "day")),
      datasets: [
        {
          label: 'Sessions',
          borderColor: 'rgba(0, 255, 0, 0.5)',
          fill: false,
          yAxisID: 'session-axis',
          data: c.session_sums.map(s => ({x: getDateStrFromUnixTime(s.bucket, "day"), y: s.count})),
        },
        {
          label: 'Page views',
          borderColor: 'rgba(0, 0, 255, 0.5)',
          fill: false,
          yAxisID: 'page-view-axis',
          data: c.pageview_sums.map(s => ({x: getDateStrFromUnixTime(s.bucket, "day"), y: s.count})),
        },
      ]
    }));
  }


}
