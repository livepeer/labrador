import Vue from 'vue';
import Vuetify from 'vuetify/lib';



Vue.use(Vuetify);

export default new Vuetify({
    theme: {
        themes: {
            dark: {
                background: '#282828',
                primary: '#00EB87',
                secondary: '#282828',
                anchor: '#00EB87'
            }
        },
        dark: true,
        options: {
            customProperties: true,
        }
    }
});