import { Component, OnInit, Input } from '@angular/core';
import { ActivatedRoute, Params } from '@angular/router';

import { Collection, BackendService } from '../backend.service';

@Component({
  selector: 'rana-collection-tracking',
  templateUrl: './tracking.component.html',
})
export class CollectionTrackingComponent implements OnInit {
  @Input() collection: Collection;
  trackingCode: string;

  constructor(
    private backend: BackendService,
    private route: ActivatedRoute,
  ) { }

  ngOnInit() {
    this.trackingCode = this.getTrackingCode();
  }

  getOrigin(): string {
    return window.location.origin;
  }

  getTrackingCode(): string {
    return `<!-- rightana -->
<script>
  (function(d, w, u, o){
    w[o]=w[o]||function(){
      (w[o].q=w[o].q||[]).push(arguments)
    };
    a=d.createElement('script'),
    m=d.getElementsByTagName('script')[0];
    a.async=1; a.src=u;
    m.parentNode.insertBefore(a,m)
  })(document, window, '${this.getOrigin()}/tracker.js', 'rightana');
  rightana('setup', '${this.getOrigin()}/api', '${this.collection.id}');
  rightana('trackPageview');
</script>
<!-- rightana -->
`;
  }

}
