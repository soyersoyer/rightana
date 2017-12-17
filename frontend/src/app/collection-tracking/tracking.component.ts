import { Component, OnInit, Input } from '@angular/core';
import { ActivatedRoute, Params } from '@angular/router';

import { Collection, BackendService } from '../backend.service';

@Component({
  selector: 'k20a-collection-tracking',
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
    return `<!-- k20a tracker -->
<script>
  (function(d, w, u, o){
    w[o]=w[o]||function(){
      (w[o].q=w[o].q||[]).push(arguments)
    };
    a=d.createElement('script'),
    m=d.getElementsByTagName('script')[0];
    a.async=1; a.src=u;
    m.parentNode.insertBefore(a,m)
  })(document, window, '${this.getOrigin()}/tracker.js', 'k20a');
  k20a('setup', '${this.getOrigin()}/api', '${this.collection.id}');
  k20a('trackPageview');
</script>
<!-- k20a tracker -->
`;
  }

}
