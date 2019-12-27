<template>
  <v-dialog v-model="activate" persistent max-width="75vw">
      <template v-slot:activator="{ on }">
          <v-btn
              large
              color="background"
              v-on="on"
            >
              <v-icon>mdi-cogs</v-icon>
            </v-btn>
      </template>
      <v-card>
          <v-card-title>
              <span class="headline">Configuration</span>
          </v-card-title>
          <v-card-text>
              <v-container>
                  <v-row>
                     <v-col cols="12" sm="6" md="6">
                         <v-text-field label="Broadcaster" v-model="setConfig.host" @change="disabled = allowUpdate()" required></v-text-field>
                     </v-col>
                     <v-col cols="12" sm="6" md="3">
                         <v-text-field label="RTMP Port" v-model="setConfig.rtmp" @change="disabled = allowUpdate()" required></v-text-field>
                     </v-col>
                     <v-col cols="12" sm="6" md="3">
                         <v-text-field label="Media Port" v-model="setConfig.media" @change="disabled = allowUpdate()" required></v-text-field>
                     </v-col>
                     <v-col cols="12">
                         <v-text-field label="File Name" v-model="setConfig.file_name" @change="disabled = allowUpdate()" required></v-text-field>
                     </v-col>
                     <v-col cols="12" sm="6" md="6">
                         <v-text-field label="Concurrent streams" v-model="setConfig.simultaneous" @change="disabled = allowUpdate()" required></v-text-field>
                     </v-col>
                     <v-col cols="12" sm="6" md="6">
                         <v-text-field label="# Renditions" v-model="setConfig.profiles_num"  @change="disabled = allowUpdate()" required></v-text-field>
                     </v-col>
                     <v-col cols="12">
                        <v-alert :value="success" transition="scale-transition" type="success">
                            Configuration updated
                        </v-alert>
                     </v-col>
                  </v-row>
              </v-container>
          </v-card-text>
          <v-card-actions>
              <v-spacer></v-spacer>
              <v-btn @click="close()" text color="background--text">
                  Cancel
              </v-btn>
              <v-btn @click="update()" dark color="primary" :disabled="disabled">Update</v-btn>
          </v-card-actions>
      </v-card>
  </v-dialog>
</template>

<script>
export default {
    data: function () {
        return {
            activate: false,
            config: {},
            setConfig: {},
            disabled: true,
            loading: false,
            success: false,
        }
    },
    async created () {
        const cfg = (await this.$http.get("http://" + process.env.VUE_APP_BASE_URL + "/config")).data
        this.config = Object.assign({}, cfg)
        this.setConfig = Object.assign({}, cfg)
    },
    methods: {
        allowUpdate: function() {
            for (let key in this.setConfig) {
                if (this.setConfig[key] != this.config[key] && this.setConfig[key] != 0) {
                    return false
                }
            }
            return true
        },
        close: function() {
            this.setConfig = Object.assign({}, this.config)
            this.activate = false
            this.success = false
            this.disabled = true
            this.loading = false
        },
        update: async function() {
            this.loading = true
            await this.$http({
                url: "http://" + process.env.VUE_APP_BASE_URL + "/config/update",
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                data: {
                    host: this.setConfig.host,
                    rtmp: typeof this.setConfig.rtmp == 'string' ? parseInt(this.setConfig.rtmp, 10) : this.setConfig.rtmp,
                    media: typeof this.setConfig.media == 'string' ? parseInt(this.setConfig.media, 10) : this.setConfig.media,
                    file_name: this.setConfig.file_name,
                    repeat: typeof this.setConfig.repeat == 'string' ? parseInt(this.setConfig.repeat, 10) : this.setConfig.repeat,
                    simultaneous: typeof this.setConfig.simultaneous == 'string' ? parseInt(this.setConfig.simultaneous, 10) : this.setConfig.simultaneous,
                    profiles_num: typeof this.setConfig.profiles_num == 'string' ? parseInt(this.setConfig.profiles_num, 10) : this.setConfig.profiles_num,
                    do_not_clear_stats: false,
                }
            })
            this.config = (await this.$http.get("http://" + process.env.VUE_APP_BASE_URL + "/config")).data
            this.setConfig = Object.assign({}, this.config)
            this.loading = false
            this.success = true
            this.disabled = true
            setTimeout(() => {
                this.success = false
            }, 3000)
        }
    },
    watch: {
        activate: function(v, o) {
            if (v && !o) this.setConfig = Object.assign({}, this.config);
        }
    }
}
</script>

<style>

</style>