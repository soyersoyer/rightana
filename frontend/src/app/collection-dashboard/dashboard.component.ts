import { Component, OnInit, OnDestroy, EventEmitter } from '@angular/core';
import { Router, ActivatedRoute, Params } from '@angular/router';

import { BackendService, AuthService, CollectionData, BucketSum } from '../backend.service';

class Interval {
  day: number;
  label: string;
}

class Bucket {
  name: string;
  label: string;
  addOne: (date: Date) => void;
  clampFromDate?: (date: Date) => void;
  clampToDate?: (date: Date) => void;
  minInterval?: Interval;
}

class DatePair {
  from: Date;
  to: Date;
}

export class Setup {
  collectionId: string;
  from: Date;
  to: Date;
  filter = {};
  events = new EventEmitter<any>();

  chartFrom: Date;
  chartTo: Date;
  chartLock = false;
  selectedBucket: number;

  emit() {
    this.events.emit(this);
  }

  set(setup: Setup) {
    this.collectionId = setup.collectionId;
    this.from = setup.from;
    this.to = setup.to;
    this.chartFrom = setup.chartFrom;
    this.chartTo = setup.chartTo;
    this.chartLock = setup.chartLock;
    this.selectedBucket = setup.selectedBucket;
    this.filter = Object.assign({}, setup.filter);
  }

  setCollectionId(collectionId: string) {
    this.collectionId = collectionId;
  }

  setDates(from: Date, to: Date) {
    this.from = from;
    this.to = to;
    this.chartFrom = from;
    this.chartTo = to;
    this.chartLock = false;
    this.selectedBucket = undefined;
    this.emit();
  }

  setDatesWOChart(from: Date, to: Date, bucket: number) {
    this.from = from;
    this.to = to;
    this.chartLock = true;
    if (this.selectedBucket !== bucket) {
      this.selectedBucket = bucket;
    } else {
      this.from = this.chartFrom;
      this.to = this.chartTo;
      this.selectedBucket = undefined;
    }
    this.emit();
  }

  add(key: string, value: string) {
    this.filter[key] = value;
    this.chartLock = false;
    this.emit();
  }

  del(key: string) {
    delete this.filter[key];
    this.chartLock = false;
    this.emit();
  }

  in(key: string): boolean {
    return this.filter[key] !== undefined;
  }
}

@Component({
  selector: 'rana-collection-dashboard',
  templateUrl: './dashboard.component.html',
})
export class CollectionDashboardComponent implements OnInit, OnDestroy {
  setup = new Setup();

  realFrom: Date;
  realTo: Date;

  intervals: Interval[] = [
    {day: 1, label: '1d'},
    {day: 7, label: '7d'},
    {day: 30, label: '30d'},
    {day: 90, label: '90d'},
    {day: 365, label: 'Year'},
  ];
  buckets: Bucket[] = [
    {name: 'hour', label: 'Hour', addOne: (date: Date) => date.setHours(date.getHours() + 1)},
    {name: 'day', label: 'Day',
      addOne: (date: Date) => date.setDate(date.getDate() + 1),
      clampFromDate: (date: Date) => {
        date.setDate(date.getDate() + 1); date.setHours(0); date.setMinutes(0); date.setSeconds(0); date.setMilliseconds(0);
      },
      clampToDate: (date: Date) => {
        date.setDate(date.getDate() + 1); date.setHours(0); date.setMinutes(0); date.setSeconds(0); date.setMilliseconds(0);
      },
      minInterval: this.intervals[1],
    },
    {name: 'week', label: 'Week',
      addOne: (date: Date) => date.setDate(date.getDate() + 7),
      clampFromDate: (date: Date) => {
        date.setDate(date.getDate() + 1); date.setHours(0); date.setMinutes(0); date.setSeconds(0); date.setMilliseconds(0);
        date.setDate(date.getDate() - (date.getDay() + 6) % 7);
      },
      clampToDate: (date: Date) => {
        date.setDate(date.getDate() + 7 - (date.getDay() + 6) % 7);
        date.setHours(0); date.setMinutes(0); date.setSeconds(0); date.setMilliseconds(0);
      },
      minInterval: this.intervals[2],
    },
    {name: 'month', label: 'Month',
      addOne: (date: Date) => date.setMonth(date.getMonth() + 1),
      clampFromDate: (date: Date) => {
        date.setDate(1);
        date.setHours(0); date.setMinutes(0); date.setSeconds(0); date.setMilliseconds(0);
      },
      clampToDate: (date: Date) => {
        date.setMonth(date.getMonth() + 1); date.setDate(1);
        date.setHours(0); date.setMinutes(0); date.setSeconds(0); date.setMilliseconds(0);
      },
      minInterval: this.intervals[3],
    },
  ];

  selectedInterval = this.intervals[1];
  selectedBucket = this.buckets[0];

  timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  collection: CollectionData;

  data: any;
  options = {
    hover: {
      onHover: (e, el) => {
        const section = el[0];
        const currentStyle = e.currentTarget.style;
        currentStyle.cursor = section ? 'pointer' : 'unset';
      }
    },
    responsive: true,
    maintainAspectRatio: false,
    scales: {
      xAxes: [{
          stacked: true
      }],
      yAxes: [{
          stacked: true,
          ticks: {
            beginAtZero: true,
          },
      }]
    },
    onClick: (c, i) => this.chartClick(c, i),
    color0: '#727b84',
    color0u: 'rgba(114, 123, 132, 0.5)',
    color1: 'rgba(0, 0, 0, 0.1)',
    color1u: 'rgba(100, 100, 100, 0.05)',
  };
  subscription: any;

  constructor(
    private backend: BackendService,
    private auth: AuthService,
    private router: Router,
    private route: ActivatedRoute,
  ) { }

  chartClick(c, i) {
    const e = i[0];
    if (e) {
      let x_value = this.data.labels[e._index];
      // to treat as local...
      if (x_value.length === 10) {
        x_value += ' 00:00';
      }
      const from = new Date(x_value);
      const to = new Date(x_value);
      this.selectedBucket.addOne(to);
      this.setup.setDatesWOChart(from, to, e._index);
      this.colorizeBuckets();
      e._chart.update();
    }
  }

  ngOnInit() {
    this.subscription = this.setup.events.subscribe(setup => {
      this.getCollectionData();
    });
    this.route.params.forEach((params: Params) => {
      this.setup.setCollectionId(params['collectionId']);
      this.today();
    });
  }

  ngOnDestroy() {
    this.subscription.unsubscribe();
  }

  today() {
    const now = new Date();
    this.realTo = new Date(now.getFullYear(), now.getMonth(), now.getDate(), now.getHours() + 1, 0, 0, 0);
    this.realFrom = this.calculateFrom(this.realTo, this.selectedInterval);
    this.setupDates();
  }

  selectInterval(interval: Interval) {
    this.selectedInterval = interval;
    this.realFrom = this.calculateFrom(this.realTo, this.selectedInterval);
    this.setupDates();
  }

  selectBucket(bucket: Bucket) {
    if (bucket.minInterval && this.selectedInterval.day < bucket.minInterval.day) {
      this.selectedInterval = bucket.minInterval;
      this.realFrom = this.calculateFrom(this.realTo, this.selectedInterval);
    }
    this.selectedBucket = bucket;
    this.setupDates();
  }

  prev() {
    this.realTo.setDate(this.realTo.getDate() - this.selectedInterval.day);
    this.realFrom = this.calculateFrom(this.realTo, this.selectedInterval);
    this.setupDates();
  }

  next() {
    this.realTo.setDate(this.realTo.getDate() + this.selectedInterval.day);
    this.realFrom = this.calculateFrom(this.realTo, this.selectedInterval);
    this.setupDates();
  }

  calculateFrom(to: Date, interval: Interval): Date {
    return new Date(
      to.getFullYear(),
      to.getMonth(),
      to.getDate() - interval.day,
      to.getHours()
    );
  }

  setupDates() {
    let from = new Date(this.realFrom);
    if (this.selectedBucket.clampFromDate) {
      this.selectedBucket.clampFromDate(from);
    }
    let to = new Date(this.realTo);
    if (this.selectedBucket.clampToDate) {
      this.selectedBucket.clampToDate(to);
    }
    this.setup.setDates(from, to);
  }

  getCollectionData() {
    if (this.setup.chartLock) {
      return;
    }
    this.backend.getCollectionData(this.setup.collectionId, this.setup.chartFrom, this.setup.chartTo, this.selectedBucket.name,
      this.timezone, this.setup.filter)
      .subscribe(collection => {
        this.collection = collection;
        this.setData(collection);
      });
  }



  padNumber(n: number): string {
    return n<10?"0"+n:""+n
  }

  getDateStrFromUnixTime(unix: number): string {
    var d = new Date(unix*1000);
    var datestr = d.getFullYear()+"-"+this.padNumber(d.getMonth()+1)+"-"+this.padNumber(d.getDate());
    if (this.selectedBucket.name === "hour") {
      return datestr+" "+this.padNumber(d.getHours())+":"+this.padNumber(d.getMinutes());
    }
    return datestr;
  }

  colorizeBuckets() {
    if (this.setup.selectedBucket !== undefined) {
      this.data.datasets[0].backgroundColor = this.collection.session_sums.map(_ => this.options.color0u)
      this.data.datasets[1].backgroundColor = this.collection.session_sums.map(_ => this.options.color1u)
      this.data.datasets[0].backgroundColor[this.setup.selectedBucket] = this.options.color0;
      this.data.datasets[1].backgroundColor[this.setup.selectedBucket] = this.options.color1;
    } else {
      this.data.datasets[0].backgroundColor = this.collection.session_sums.map(_ => this.options.color0)
      this.data.datasets[1].backgroundColor = this.collection.session_sums.map(_ => this.options.color1)
    }
  }

  setData(collection: CollectionData) {
    const labels = collection.session_sums.map(s => this.getDateStrFromUnixTime(s.bucket));
    const datasets = [
      {
        label: 'Sessions',
        data: collection.session_sums.map(s => ({x: this.getDateStrFromUnixTime(s.bucket), y: s.count})),
      },
      {
        label: 'Page views',
        data: collection.pageview_sums.map(s => ({x: this.getDateStrFromUnixTime(s.bucket), y: s.count})),
      },
    ];
    this.data = {labels, datasets};
    this.colorizeBuckets();
  }

  showChart(): boolean {
    return !this.router.url.endsWith('settings');
  }


  keys(o: any): string[] {
    return Object.keys(o);
  }
}
