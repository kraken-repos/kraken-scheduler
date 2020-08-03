
package com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction;

import com.fasterxml.jackson.annotation.*;
import lombok.Data;

import java.util.HashMap;
import java.util.Map;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
@JsonPropertyOrder({
    "name",
    "value",
    "confidence",
    "page",
    "coordinates"
})
public class LineItem {

    @JsonProperty("name")
    private String name;
    @JsonProperty("value")
    private String value;
    @JsonProperty("confidence")
    private Double confidence;
    @JsonProperty("page")
    private Long page;
    @JsonProperty("coordinates")
    private Coordinates coordinates;
    @JsonIgnore
    private Map<String, Object> additionalProperties = new HashMap<String, Object>();
}
