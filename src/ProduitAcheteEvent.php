<?php

class ProduitAcheteEvent implements DomainEventInterface
{
    const TYPE = 'produit.achete';

    private $rootEntityId;
    private $occuredOn;
    private $informations;

    public function __construct($produitId, $clientId)
    {
        $this->rootEntityId = $produitId;
        $this->occuredOn = new \DateTime();
        $this->informations = ['clientId' => $clientId];
    }

    public function getType()
    {
        return self::TYPE;
    }

    public function getRootEntityId()
    {
        return $this->rootEntityId;
    }

    public function occurredOn()
    {
        return $this->occuredOn;
    }

    public function getEventInformationsAsArray()
    {
        return $this->informations;
    }
}

// Publie le domain event
$event = new ProduitAcheteEvent($produit->getId(), $client->getId());

$dispatcher = new DomainEventDispatcher;
$dispatcher->dispatch($event);