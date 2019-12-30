<template>
  <v-app id="inspire">
    <v-content>
      <v-container
        class="fill-height"
        fluid
      >
        <v-data-iterator
          :items="items"
          :items-per-page.sync="itemsPerPage"
          :sort-by="parseKey(sortBy)"
          :sort-desc="sortDesc"
          :loading="this.loadingTable"
        >

      <template v-slot:header>
        <v-toolbar
          dark
          color="primary"
          class="mb-1 pl-3"
        >
          <div class="row align-center">
            <img width="50" height="50" src="./assets/lp.svg" />
            <span class="display-1 ml-2 secondary--text">LABRADOR - </span>
            <span class="headline ml-2 secondary--text"><span class="underline">L</span>ivepeer <span class="underline">A</span>utomated <span class="underline">Br</span>oadcast <span class="underline">A</span>n<span class="underline">d</span> Monit<span class="underline">or</span></span> 
          </div>

          <template v-if="$vuetify.breakpoint.mdAndUp">
            <v-spacer></v-spacer>
            <v-select
              class="secondary"
              dark
              v-model="sortBy"
              flat
              solo-inverted
              hide-details
              :items="keys"
              prepend-inner-icon="mdi-magnify"
              label="Sort by"
            ></v-select>
            <v-spacer></v-spacer>
            <v-btn-toggle
              v-model="sortDesc"
              mandatory
            >
              <v-btn
                large
                depressed
                color="background"
                :value="false"
              >
                <v-icon>mdi-arrow-up</v-icon>
              </v-btn>
              <v-btn
                large
                depressed
                color="background"
                :value="true"
              >
                <v-icon>mdi-arrow-down</v-icon>
              </v-btn>
            </v-btn-toggle>
            <v-spacer></v-spacer>
            <config-dialog class="pr-3" :config="config" />
          </template>

        </v-toolbar>

      </template>

        <template v-slot:default="props">
        <v-row>
          <v-col
            v-for="item in props.items"
            :key="item.base_manifest_id"
            cols="12"
            sm="6"
            md="4"
            lg="3"
          >
            <v-card>
              <v-card-title>
                <span class="primary--text overline font-weight-bold" style="font-size:14px !important;">
                  {{ item.base_manifest_id }}
                </span>
              </v-card-title>

              <v-divider></v-divider>

              <v-list dense>

                <v-list-item>
                  <v-list-item-content style="font-size:11px !important;" class="overline" :class="{ 'primary--text': sortBy === formatKey('start_time') }">Start time</v-list-item-content>
                  <v-list-item-content style="font-size:11px !important;" class="overline align-end" :class="{ 'primary--text': sortBy === formatKey('Start time') }">{{$moment(item.start_time).format('MMM Do YYYY HH:mm')}}</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content style="font-size:11px !important;" class="overline" :class="{ 'primary--text': sortBy === formatKey('success_rate') }">Success rate</v-list-item-content>
                  <v-list-item-content style="font-size:11px !important;" class="overline align-end" :class="{ 'primary--text': sortBy === formatKey('success_rate') }">{{item.success_rate}} %</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content style="font-size:11px !important;" class="overline" :class="{ 'primary--text': sortBy === formatKey('avg_latency') }">Avg. latency</v-list-item-content>
                  <v-list-item-content style="font-size:11px !important;" class="overline align-end" :class="{ 'primary--text': sortBy === formatKey('avg_latency') }">{{$moment.duration(item.avg_latency*Math.pow(10, -6)).asSeconds().toString().slice(0,4)}} seconds</v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content  style="font-size:11px !important;" class="overline">
                    Concurrent Streams
                  </v-list-item-content>
                  <v-list-item-content style="font-size:11px !important;" class="overline align-end" >
                    {{item.rtmp_streams}}
                  </v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content  style="font-size:11px !important;" class="overline">
                    # Renditions
                  </v-list-item-content>
                  <v-list-item-content style="font-size:11px !important;" class="overline align-end" >
                    {{item.profiles_num}}
                  </v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content  style="font-size:11px !important;" class="overline">
                    Segments uploaded
                  </v-list-item-content>
                  <v-list-item-content style="font-size:11px !important;" class="overline align-end" >
                    {{item.sent_segments}} / {{item.total_segments_to_send}}
                  </v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content  style="font-size:11px !important;" class="overline">
                    Segments downloaded
                  </v-list-item-content>
                  <v-list-item-content style="font-size:11px !important;" class="overline align-end" >
                    {{item.downloaded_segments}} / {{item.should_have_downloaded_segments}}
                  </v-list-item-content>
                </v-list-item>

                <v-list-item>
                  <v-list-item-content style="font-size:11px !important;" class="overline" :class="{ 'primary--text': sortBy === formatKey('finished') }">Finished</v-list-item-content>
                  <v-list-item-content style="font-size:11px !important;" class="overline align-end" :class="{ 'primary--text': sortBy === formatKey('finished') }">{{item.finished}}</v-list-item-content>
                </v-list-item>

              </v-list>
            </v-card>
          </v-col>
        </v-row>
      </template>
        </v-data-iterator>
      </v-container>
    </v-content>

    <v-footer app  class="row justify-center">
      <div class="row justify-space-around">
        <div>
          <span>Livepeer inc. &copy; 2019</span>
        </div>
        <div>
          <a class="link" href="https://www.livepeer.org" target="_blank">
            www.livepeer.org
          </a>
        </div>
      </div>
    </v-footer>

  </v-app>
</template>

<script>
import Config from './components/Config.vue'
export default {
    props: {
      source: String,
    },
    data: () => ({
      loadingTable: false,
      sortBy: 'Start time',
      sortDesc: true,
      keys: [
        'Start time',
        'Success rate',
        'Finished',
        'Avg latency'
      ],
      itemsPerPage: 10,
      items: [],
      config: {},
    }),
    methods: {
      parseKey: (key) => {
        let k = key
        k = k.toLowerCase()
        return k.replace(" ", "_")
      },
      formatKey: (key) => {
        let k = key
        k = k.replace("_", " ")
        let first = k.slice(0, 1)
        k = k.slice(1)
        first = first.toUpperCase()
        return first + k
      },
      formatStats: (stats) => {
        let items = []
        for (const key in stats) {
          let item = stats[key]
          item["base_manifest_id"] = key
          item["avg_latency"] = item.transcoded_latencies.avg
          items.push(item)
        }
        return items
      }
    },
    async created() {
      this.loadingTable = true
      this.items = this.formatStats((await this.$http.get("http://" + process.env.VUE_APP_BASE_URL + "/stats/all")).data)
      this.loadingTable = false
      setInterval(async () => {
         this.items = this.formatStats((await this.$http.get("http://" + process.env.VUE_APP_BASE_URL + "/stats/all")).data)
      }, 30000)
    },
    components: {
        "config-dialog": Config
    }
  }

</script>

<style>

body {
  background: var(--v-background-base) !important;
}

.v-application {
      background-color: var(--v-background-base) !important;
}

.v-data-iterator {
  width: 100%;
  height: 100%;
  position: relative;
}

.v-data-footer {
  position: absolute;
  bottom: 0;
    width: 100%;
}

.link {
  text-decoration: none;
}

.underline {
  text-decoration: underline;
}

</style>