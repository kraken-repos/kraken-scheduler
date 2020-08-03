package com.sap.s4hana.eureka.business.krakenscheduler;

import com.sap.s4hana.eureka.framework.event.annotation.EnableConsumer;
import com.sap.s4hana.eureka.framework.rest.swagger.SwaggerAutoConfigruation;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@EnableConsumer
@SwaggerAutoConfigruation
@SpringBootApplication(scanBasePackages = {
        "com.sap.s4hana.eureka.business.krakenscheduler",
        "com.sap.s4hana.eureka.framework"
})
public class Application {

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }

}
