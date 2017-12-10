<?php

class FacturationProduitAcheteListener implements DomainEventListenerInterface
{
    private $facturationService;

    public function __construct(FacturationService $facturationService)
    {
        $this->facturationService = $facturationService;
    }

    public function handle(DomainEvent $event)
    {
        // Edite la facture pour l'achat du produit
    }
}

// Enregistre le listener
$facturationService = // Service de facturation...
$listener = new FacturationProduitAcheteListener($facturationService);

$dispatcher = new DomainEventDispatcher;
$dispatcher->addListener('produit.achete', $listener);