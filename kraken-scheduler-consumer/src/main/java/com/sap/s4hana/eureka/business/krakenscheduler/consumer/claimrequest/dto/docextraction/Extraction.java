
package com.sap.s4hana.eureka.business.krakenscheduler.consumer.claimrequest.dto.docextraction;

import com.fasterxml.jackson.annotation.*;
import lombok.Data;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Data
@JsonInclude(JsonInclude.Include.NON_NULL)
public class Extraction {

    @JsonProperty("headerFields")
    private List<HeaderField> headerFields = null;
    @JsonProperty("lineItems")
    private List<List<LineItem>> lineItems = null;
    @JsonIgnore
    private Map<String, Object> additionalProperties = new HashMap<String, Object>();
}
